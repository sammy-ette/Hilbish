-- Tests for the Paperbush classifier.
--
-- Run under Hilbish from the repo root:
--   hilbish nature/paperbush/classifier_test.lua
-- Exits 0 if all assertions pass, 1 otherwise.

package.path = './?.lua;' .. package.path
local classifier = require('nature.paperbush.classifier')

local total, failures = 0, 0
local function fail(name, detail)
	failures = failures + 1
	io.stderr:write(string.format('FAIL: %s%s\n', name, detail and ('\n      ' .. detail) or ''))
end

local function eqList(a, b)
	if #a ~= #b then return false end
	for i = 1, #a do if a[i] ~= b[i] then return false end end
	return true
end

-- expectTypes asserts the tags of the top-level statements of `src`.
local function expectTypes(src, want)
	total = total + 1
	local got = {}
	for _, s in ipairs(classifier.parse(src)) do got[#got + 1] = s.type end
	if not eqList(got, want) then
		fail('types ' .. ('%q'):format(src), 'got {' .. table.concat(got, ', ') .. '}')
	end
end

-- check is a generic boolean assertion.
local function check(name, cond, detail)
	total = total + 1
	if not cond then fail(name, detail) end
end

-- bodyOf returns the first `body` part of a block node.
local function bodyOf(stmt)
	for _, p in ipairs(stmt.parts) do if p.body then return p.body end end
	return {}
end

-- simple statement tagging ------------------------------------------------

expectTypes('ls -la', { 'command' })
expectTypes('git status', { 'command' })
expectTypes('local n = 0', { 'lua' })
expectTypes("print 'hi'", { 'dispatch' })  -- global callable, resolved at runtime
expectTypes('echo "hi"', { 'dispatch' })   -- not callable at runtime => command
expectTypes('grep "foo" f.lua', { 'command' }) -- trailing token => not valid Lua sugar
expectTypes('x = a | b', { 'lua' })               -- bitwise-or, not a pipeline
expectTypes("local f = 1; f 'y'", { 'lua', 'lua' }) -- f known local => sugar call
expectTypes("greet 'x'", { 'dispatch' })          -- unknown head + string => runtime dispatch

do -- dispatch records the head name
	local s = classifier.parse("greet 'x'")[1]
	check('dispatch head', s.type == 'dispatch' and s.head == 'greet', 'head=' .. tostring(s.head))
end

-- pipelines ---------------------------------------------------------------

expectTypes('ls | grep x', { 'pipeline' })

do -- mixed pipeline splits into shell + lua stages
	local s = classifier.parse('ls | @(string.upper) | grep x')[1]
	local ok = s.type == 'pipeline' and #s.stages == 3
		and s.stages[1].kind == 'shell'
		and s.stages[2].kind == 'lua' and s.stages[2].expr == 'string.upper'
		and s.stages[3].kind == 'shell'
	check('pipeline stages', ok, s.type == 'pipeline' and (#s.stages .. ' stages') or s.type)
end

-- a string arg before the pipe must not short-circuit to command (rule 4 order)
expectTypes('echo "hi" | @(string.upper)', { 'pipeline' })

do -- a bare `function ... end` in pipe position is a lua stage
	local s = classifier.parse('ls | function(x) return x end')[1]
	check('bare function stage',
		s.type == 'pipeline' and #s.stages == 2 and s.stages[2].kind == 'lua', s.type)
end

-- blocks: commands nest, multi-line, multi-clause ------------------------

expectTypes('for i = 1, 3 do ls end', { 'block' })
expectTypes('local f = function() return 1 end', { 'lua' })                 -- anon literal is one simple stmt
expectTypes('local f = function() if x then return 1 end end', { 'lua' })   -- nested end inside anon literal

do -- command nested in a for-block
	local s = classifier.parse('for i = 1, 3 do ls end')[1]
	local b = s.type == 'block' and bodyOf(s) or {}
	check('for body is command', #b == 1 and b[1].type == 'command', '#=' .. #b)
end

do -- multi-line block stays intact, body command
	local s = classifier.parse('for i = 1, 2 do\n  echo @(i)\nend')[1]
	local b = s.type == 'block' and bodyOf(s) or {}
	check('multiline for body', s.type == 'block' and #b == 1 and b[1].type == 'command')
end

do -- flagship: for ... in $(...) do <command> end
	local s = classifier.parse('for f in $(ls) do convert @(f) end')[1]
	local b = s.type == 'block' and bodyOf(s) or {}
	check('for-in block command', s.type == 'block' and #b == 1 and b[1].type == 'command')
end

do -- if/elseif/else => three bodies
	local s = classifier.parse('if a then x() elseif b then y() else z() end')[1]
	local bodies = 0
	if s.type == 'block' then for _, p in ipairs(s.parts) do if p.body then bodies = bodies + 1 end end end
	check('if/elseif/else bodies', s.type == 'block' and bodies == 3, 'bodies=' .. bodies)
end

do -- nested blocks: if { for { command } }
	local s = classifier.parse('if x then for i = 1, 2 do ls end end')[1]
	local inner = bodyOf(s)[1]
	local b2 = inner and inner.type == 'block' and bodyOf(inner) or {}
	check('nested block command', inner and inner.type == 'block' and #b2 == 1 and b2[1].type == 'command')
end

-- fuzz regressions ------------------------------------------------------

-- pipeline routing: a NAME (command / dotted-ref) head -> snail pipeline;
-- a value/sigil head -> all-Lua value pipeline (lua/thru).
expectTypes('ls | grep x', { 'pipeline' })
expectTypes("string.match 'x' | grep y", { 'pipeline' }) -- dotted NAME head
expectTypes('$PATH | string.split ":"', { 'lua' })       -- $VAR head -> value pipe
expectTypes('@(f) | string.upper', { 'lua' })            -- @() head -> value pipe
expectTypes('$(cmd) | tonumber', { 'lua' })              -- $() head -> value pipe

-- stray block terminators must not hang the parser
do
	for _, s in ipairs({ 'end', 'else', 'until', 'elseif', 'end end then' }) do
		check('no-hang ' .. ('%q'):format(s), pcall(classifier.parse, s))
	end
end

-- summary -----------------------------------------------------------------

if failures == 0 then
	print(string.format('ok - %d assertions passed', total))
	os.exit(0)
end

io.stderr:write(string.format('%d/%d assertions failed\n', failures, total))
os.exit(1)
