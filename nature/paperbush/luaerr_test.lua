-- Tests for luaerr: mapping load()/pcall() errors back to source.
--
-- These run the REAL codegen -> load()/pcall() pipeline (not hand-built error
-- strings), because the exact wording/positions Lua gives are runtime-
-- specific (verified directly against this Hilbish build: load() syntax
-- errors carry line:col, pcall() runtime errors carry only a line and never
-- name the offending variable).
--
-- Run under Hilbish from the repo root:
--   hilbish nature/paperbush/luaerr_test.lua
-- Exits 0 if all assertions pass, 1 otherwise.

package.path = './?.lua;' .. package.path

local calls = {}
package.loaded['nature.paperbush.runtime'] = {
	exec = function(c) calls[#calls + 1] = { 'exec', c }; return 0 end,
}

local codegen = require('nature.paperbush.codegen')
local luaerr = require('nature.paperbush.luaerr')

local total, failures = 0, 0
local function fail(name, detail)
	failures = failures + 1
	io.stderr:write(string.format('FAIL: %s%s\n', name, detail and ('\n      ' .. detail) or ''))
end
local function check(name, cond, detail)
	total = total + 1
	if not cond then fail(name, detail) end
end

-- runLoad/runRuntime drive `src` through the real pipeline and return the
-- luaerr-mapped diagnostic for whichever failure actually happens.
local function runLoad(src)
	local code = codegen.chunk(src)
	local map = codegen.lastMap()
	local f, err = load(code, '=paperbush')
	if f then return nil end
	return luaerr.fromLoad(err, src, map)
end

local function runRuntime(src)
	local code = codegen.chunk(src)
	local map = codegen.lastMap()
	local f = assert(load(code, '=paperbush'))
	local ok, err = pcall(f)
	if ok then return nil end
	return luaerr.fromRuntime(err, src, map)
end

-- load() errors: gopher-lua gives line:col, disambiguated against the span --

do -- two '=' on the line: the SECOND one is the actual error, not the first
	local src = 'local x = = 5'
	local d = runLoad(src)
	check('double-= produces a diagnostic', d ~= nil)
	if d then
		check('double-= code', d.code == 'syntax', tostring(d.code))
		check('double-= label exists', d.labels[1] ~= nil)
		if d.labels[1] then
			check('double-= points at the SECOND =, not the first', d.labels[1].line == 1 and d.labels[1].col == 11,
				'line=' .. tostring(d.labels[1].line) .. ' col=' .. tostring(d.labels[1].col))
		end
	end
end

do -- same error on line 2, so there's no chunk-prefix column shift to correct
	local src = 'ls -la\nlocal x = = 5'
	local d = runLoad(src)
	check('double-= line2 diagnostic', d ~= nil)
	if d and d.labels[1] then
		check('double-= line2 position', d.labels[1].line == 2 and d.labels[1].col == 11,
			'line=' .. tostring(d.labels[1].line) .. ' col=' .. tostring(d.labels[1].col))
	end
end

-- pcall() runtime errors: no locator text or column exists, so these resolve
-- to the whole responsible STATEMENT, not a sub-span --

do
	local src = 'foo(1, 2)'
	local d = runRuntime(src)
	check('nil-call produces a diagnostic', d ~= nil)
	if d then
		check('nil-call code', d.code == 'call-nil', tostring(d.code))
		check('nil-call message has no fabricated name', d.message == 'attempt to call a nil value', d.message)
		check('nil-call points at the whole statement', d.labels[1]
			and src:sub(d.labels[1].col, d.labels[1].col + d.labels[1].len - 1) == 'foo(1, 2)',
			d.labels[1] and (tostring(d.labels[1].col) .. '+' .. tostring(d.labels[1].len)) or 'no label')
	end
end

do -- the SECOND of two statements fails - position must land on line 2, not 1
	local src = 'ls -la\nfoo(1, 2)'
	local d = runRuntime(src)
	check('second-statement nil-call diagnostic', d ~= nil)
	if d and d.labels[1] then
		check('second-statement line', d.labels[1].line == 2, tostring(d.labels[1].line))
		check('second-statement spans the right statement',
			d.labels[1].col == 1 and d.labels[1].len == #'foo(1, 2)',
			'col=' .. tostring(d.labels[1].col) .. ' len=' .. tostring(d.labels[1].len))
	end
end

-- generic fallback: an unrecognized message shape still frames cleanly ------

do
	local d = luaerr.fromLoad('=paperbush:1: some unrecognized future error shape', 'whatever', {})
	check('unrecognized shape falls back to message-only', d ~= nil and d.message == 'some unrecognized future error shape')
	check('unrecognized shape has no labels (no anchor)', d ~= nil and #d.labels == 0)
end

do -- non-string error values (e.g. error({...})) don't crash the mapper
	local d = luaerr.fromRuntime({ code = 42 }, 'whatever', {})
	check('non-string error handled', d ~= nil and d.severity == 'error')
end

do -- a line with no matching anchor degrades to message-only, not a crash
	local d = luaerr.fromLoad('=paperbush:99: attempt to call a nil value', 'src', {})
	check('no-anchor line degrades gracefully', d ~= nil and #d.labels == 0)
end

-- summary -------------------------------------------------------------------

if failures == 0 then
	print(string.format('ok - %d assertions passed', total))
	os.exit(0)
end

io.stderr:write(string.format('%d/%d assertions failed\n', failures, total))
os.exit(1)
