-- Tests for the Paperbush lexer.
--
-- Run under Hilbish from the repo root:
--   hilbish nature/paperbush/lexer_test.lua
-- Exits 0 if all assertions pass, 1 otherwise.

package.path = './?.lua;' .. package.path
local lexer = require('nature.paperbush.lexer')
local k = lexer.kinds

local total, failures = 0, 0

local function fail(name, detail)
	failures = failures + 1
	io.stderr:write(string.format('FAIL: %s%s\n', name, detail and ('\n      ' .. detail) or ''))
end

-- non-EOF tokens for `src`
local function lex(src)
	local out = {}
	for _, t in ipairs(lexer.tokenize(src)) do
		if t.kind ~= k.EOF then out[#out + 1] = t end
	end
	return out
end

local function render(toks)
	local out = {}
	for _, t in ipairs(toks) do out[#out + 1] = t.kind .. '(' .. t.value .. ')' end
	return table.concat(out, ' ')
end

-- eq asserts the token kinds (and values, when given) for `src`.
-- expected is a list of { KIND } or { KIND, value }.
local function eq(src, expected)
	total = total + 1
	local toks = lex(src)
	local ok = #toks == #expected
	if ok then
		for i, e in ipairs(expected) do
			if toks[i].kind ~= e[1] or (e[2] ~= nil and toks[i].value ~= e[2]) then
				ok = false
				break
			end
		end
	end
	if not ok then fail('lex ' .. ('%q'):format(src), 'got: ' .. render(toks)) end
end

local function anyIncomplete(src)
	for _, t in ipairs(lexer.tokenize(src)) do
		if t.incomplete then return true end
	end
	return false
end

-- incomplete asserts whether `src` contains an unterminated construct.
local function incomplete(src, want)
	total = total + 1
	if anyIncomplete(src) ~= want then
		fail('incomplete ' .. ('%q'):format(src), 'expected ' .. tostring(want))
	end
end

-- token kinds -------------------------------------------------------------

eq('ls -la', { { k.NAME, 'ls' }, { k.OP, '-' }, { k.NAME, 'la' } })
eq('local n = 0', { { k.KEYWORD, 'local' }, { k.NAME, 'n' }, { k.ASSIGN, '=' }, { k.NUMBER, '0' } })
eq("print 'hi'", { { k.NAME, 'print' }, { k.STRING, 'hi' } })
eq('t[1]', { { k.NAME, 't' }, { k.LBRACKET, '[' }, { k.NUMBER, '1' }, { k.RBRACKET, ']' } })
eq('a >= b', { { k.NAME, 'a' }, { k.OP, '>=' }, { k.NAME, 'b' } })
eq('git status', { { k.NAME, 'git' }, { k.NAME, 'status' } })
eq('cd "my dir"; echo done', {
	{ k.NAME, 'cd' }, { k.STRING, 'my dir' }, { k.SEMI, ';' },
	{ k.NAME, 'echo' }, { k.NAME, 'done' },
})

-- strings & comments ------------------------------------------------------

eq('echo "a $(b)"', { { k.NAME, 'echo' }, { k.STRING, 'a $(b)' } }) -- $( inside string stays text
eq('-- $(x)', { { k.COMMENT, ' $(x)' } })                          -- comment, no operator
eq('[[long]]', { { k.LONGSTRING, 'long' } })
eq('[==[ a ]==]', { { k.LONGSTRING, ' a ' } })

-- sigils (value = inner text, balanced) -----------------------------------

eq('local b = $(git status)', {
	{ k.KEYWORD, 'local' }, { k.NAME, 'b' }, { k.ASSIGN, '=' }, { k.CAPTURE, 'git status' },
})
eq('echo $(date) | @(string.upper)', {
	{ k.NAME, 'echo' }, { k.CAPTURE, 'date' }, { k.PIPE, '|' }, { k.EVAL, 'string.upper' },
})
eq('x = @(f($(whoami)))', {
	{ k.NAME, 'x' }, { k.ASSIGN, '=' }, { k.EVAL, 'f($(whoami))' }, -- nesting balanced
})
eq('run $[ls -la]', { { k.NAME, 'run' }, { k.RUN, 'ls -la' } })
eq('echo $HOME', { { k.NAME, 'echo' }, { k.ENV, 'HOME' } })
eq('echo ${HOME}', { { k.NAME, 'echo' }, { k.ENV, 'HOME' } })
eq('grep $(echo ")") file', { -- ) inside string must not close early
	{ k.NAME, 'grep' }, { k.CAPTURE, 'echo ")"' }, { k.NAME, 'file' },
})

-- incomplete detection ----------------------------------------------------

incomplete('echo "all good"', false)
incomplete('command ""invalid""', false)
incomplete('command "invalid""', true)  -- trailing lone "
incomplete("command 'invalid\"", true)  -- embedded " inside unterminated '
incomplete('local b = $(git status)', false)
incomplete('local b = $(git status', true)  -- unclosed $(
incomplete('run $[ls -la', true)            -- unclosed $[
incomplete('echo ${HOME', true)             -- unclosed ${
incomplete('x = [[unclosed', true)          -- unclosed long string
incomplete('local x = 1 + 2', false)

-- `--` is a comment only before whitespace/EOL/`[`; `--flag` stays a flag (fuzz regression)
eq('-- a comment', { { k.COMMENT } })
eq('--[[ b ]]', { { k.COMMENT } })
eq('ls --all', { { k.NAME, 'ls' }, { k.OP, '--' }, { k.NAME, 'all' } })

-- summary -----------------------------------------------------------------

if failures == 0 then
	print(string.format('ok - %d assertions passed', total))
	os.exit(0)
end

io.stderr:write(string.format('%d/%d assertions failed\n', failures, total))
os.exit(1)
