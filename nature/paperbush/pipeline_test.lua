-- Tests for the Paperbush pipeline executor (M.pipeline), against the REAL
-- runtime. Run under Hilbish from the repo root:
--   hilbish nature/paperbush/pipeline_test.lua
-- Exits 0 if all pass, 1 otherwise.

package.path = './?.lua;' .. package.path
local M = require('nature.paperbush.runtime')
local codegen = require('nature.paperbush.codegen')

local total, failures = 0, 0
local function fail(n, d)
	failures = failures + 1
	io.stderr:write('FAIL: ' .. n .. (d and ('\n      ' .. d) or '') .. '\n')
end
local function check(name, cond, detail)
	total = total + 1
	if not cond then fail(name, detail) end
end

-- run `stages` and return (capturedInput, exitCode). A capturing lua stage is
-- appended so the final stage's output is grabbed instead of hitting the
-- terminal; it doesn't touch the exit code (only shell stages do), so the last
-- real shell stage's exit is preserved.
local captured
local function run(stages)
	local s = {}
	for _, st in ipairs(stages) do s[#s + 1] = st end
	s[#s + 1] = { lua = function(input) captured = input; return '' end }
	captured = nil
	local exit = M.pipeline(s)
	return captured, exit
end

-- shell -> capture
do
	local out = run({ { shell = 'echo hi' } })
	check('shell stage', out == 'hi\n', 'got ' .. ('%q'):format(tostring(out)))
end

-- shell -> shell (real pipe) -> capture. `tr` is external and reads stdin
-- (the `cat` builtin does not).
do
	local out = run({ { shell = 'echo hello' }, { shell = 'tr a-z A-Z' } })
	check('shell|shell', out == 'HELLO\n', 'got ' .. ('%q'):format(tostring(out)))
end

-- lua -> shell: a lua stage seeds the next command's stdin
do
	local out = run({ { lua = function() return 'seed\n' end }, { shell = 'tr a-z A-Z' } })
	check('lua|shell', out == 'SEED\n', 'got ' .. ('%q'):format(tostring(out)))
end

-- shell -> lua passthrough -> shell: output flows across both boundaries
do
	local out = run({
		{ shell = 'echo hi' },
		{ lua = function(s) return s end },
		{ shell = 'tr a-z A-Z' },
	})
	check('shell|lua|shell', out == 'HI\n', 'got ' .. ('%q'):format(tostring(out)))
end

-- exit code is the last shell stage's
do
	local _, e1 = run({ { shell = 'false' } })
	check('exit nonzero', e1 ~= 0, 'got ' .. tostring(e1))
	local _, e2 = run({ { shell = 'true' } })
	check('exit zero', e2 == 0, 'got ' .. tostring(e2))
end

-- end-to-end: source -> codegen -> run against the real runtime. The final lua
-- stage stashes its input in a global so we can assert it without terminal I/O.
do -- a string arg before the pipe (regression: must stay a pipeline)
	_G.PB_CAP = nil
	assert(load(codegen.chunk([[printf 'a\nb\n' | @(function(s) _G.PB_CAP = (s:gsub('%s+$', '')):upper() end)]])))()
	check('e2e string-arg pipeline', _G.PB_CAP == 'A\nB', 'got ' .. ('%q'):format(tostring(_G.PB_CAP)))
end

do -- a bare function literal in pipe position (regression)
	_G.PB_CAP = nil
	assert(load(codegen.chunk([[printf 'x\n' | function(s) _G.PB_CAP = (s:gsub('%s+$', '')):upper() end]])))()
	check('e2e bare-function pipeline', _G.PB_CAP == 'X', 'got ' .. ('%q'):format(tostring(_G.PB_CAP)))
	_G.PB_CAP = nil
end

if failures == 0 then
	print(('ok - %d assertions passed'):format(total))
	os.exit(0)
end
io.stderr:write(('%d/%d assertions failed\n'):format(failures, total))
os.exit(1)
