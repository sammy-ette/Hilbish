-- Tests for the Paperbush codegen.
--
-- Run under Hilbish from the repo root:
--   hilbish nature/paperbush/codegen_test.lua
-- Exits 0 if all pass, 1 otherwise. The runtime is stubbed (via package.loaded)
-- so this checks codegen output and behavior, not real command execution.

package.path = './?.lua;' .. package.path

-- stub `paper` so generated chunks run and record what they would do
local calls = {}
package.loaded['nature.paperbush.runtime'] = {
	exec     = function(c) calls[#calls + 1] = { 'exec', c }; return 0 end,
	capture  = function(c) calls[#calls + 1] = { 'capture', c }; return 'CAP' end,
	result   = function(c) calls[#calls + 1] = { 'result', c }; return { exitCode = 0 } end,
	env      = function(v) return 'E:' .. v end,
	dispatch = function(n, a, c) calls[#calls + 1] = { 'dispatch', n, a[1], c }; return 0 end,
	quote    = function(v) return 'Q[' .. tostring(v) .. ']' end,
	pipeline = function(stages) calls[#calls + 1] = { 'pipeline', stages }; return 0 end,
}

local codegen = require('nature.paperbush.codegen')

local total, failures = 0, 0
local function fail(n, d)
	failures = failures + 1
	io.stderr:write('FAIL: ' .. n .. (d and ('\n      ' .. d) or '') .. '\n')
end

local function eq(src, expected)
	total = total + 1
	local got = codegen.rewrite(src)
	if got ~= expected then
		fail('rewrite ' .. ('%q'):format(src), 'want: ' .. expected .. '\n      got:  ' .. got)
	end
end

local function valid(src)
	total = total + 1
	local ok, err = load(codegen.chunk(src))
	if not ok then fail('valid ' .. ('%q'):format(src), tostring(err)) end
end

local function check(name, cond, detail)
	total = total + 1
	if not cond then fail(name, detail) end
end

-- exact codegen -----------------------------------------------------------

eq('local n = 0', 'local n = 0')
eq('ls -la', 'paper.exec("ls -la")')
eq('local b = $(git status)', 'local b = paper.capture("git status")')
eq('echo $HOME', 'paper.exec("echo $HOME")')              -- $VAR left for snail in a command
eq('print($USER)', 'print(paper.env("USER"))')            -- $VAR rewritten in Lua
eq('ls @(dir)', 'paper.exec("ls " .. paper.quote(dir))')  -- @() spliced into a command
eq("greet 'x'", 'paper.dispatch("greet", { \'x\' }, "greet \'x\'")')
eq('ls | @(string.upper) | grep x',
	'paper.pipeline({ { shell = "ls " }, { lua = string.upper }, { shell = "grep x" } })')

-- generated Lua is syntactically valid ------------------------------------

valid('for i = 1, 3 do ls end')
valid('if x then print(x) else ls end')
valid('for f in $(ls):gmatch("%S+") do convert @(f) end')
valid("local f = function() return 1 end")

-- behavior of the generated chunk (against the stub) ----------------------

do
	calls = {}
	load(codegen.chunk("local d = 'x'; ls @(d)"))()
	local c = calls[1]
	check('splice exec', c and c[1] == 'exec' and c[2] == 'ls Q[x]',
		c and (c[1] .. ' ' .. tostring(c[2])) or 'no call')
end

do
	calls = {}
	load(codegen.chunk("greet 'hi'"))()
	local c = calls[1]
	check('dispatch call', c and c[1] == 'dispatch' and c[2] == 'greet' and c[3] == 'hi', c and c[1])
end

do
	calls = {}
	load(codegen.chunk('ls | @(string.upper)'))()
	local c = calls[1]
	local ok = c and c[1] == 'pipeline' and #c[2] == 2
		and c[2][1].shell == 'ls ' and c[2][2].lua == string.upper
	check('pipeline call', ok, c and ('stages=' .. #c[2]) or 'no call')
end

do
	calls = {}
	load(codegen.chunk('for i = 1, 2 do ls end'))()
	check('block loops command', #calls == 2 and calls[1][1] == 'exec' and calls[2][1] == 'exec',
		'#calls=' .. #calls)
end

do -- a string arg before the pipe still produces a pipeline (rule 4 order)
	calls = {}
	load(codegen.chunk('echo "hi" | @(string.upper)'))()
	local c = calls[1]
	check('string-arg pipe -> pipeline', c and c[1] == 'pipeline' and #c[2] == 2
		and c[2][2].lua == string.upper, c and c[1] or 'no call')
end

do -- a bare function literal becomes a lua stage
	calls = {}
	load(codegen.chunk('ls | function(s) return s end'))()
	local c = calls[1]
	check('bare function -> lua stage', c and c[1] == 'pipeline' and #c[2] == 2
		and type(c[2][2].lua) == 'function', c and c[1] or 'no call')
end

-- commands recurse into anonymous function literal bodies
valid('local f = function() ls end')
valid('ls | @(function(l) convert @(l) end)')
do
	local out = codegen.rewrite('local f = function() ls end')
	check('anon fn body recurses', out:find('paper.exec', 1, true) ~= nil, out)
end
do
	local out = codegen.rewrite('map(xs, function(x) build @(x) end)')
	check('anon fn arg recurses', out:find('paper.quote(x)', 1, true) ~= nil, out)
end

-- fuzz regressions: these well-formed programs must compile
valid('if function() if y then z end end then ls end')   -- nested then in condition
valid('local f = function() ls --all $(wc foo) end')     -- --flag inside a function body
valid("string.match 'x' | grep y")                       -- functional pipe head
valid('@(f) | grep y | string.upper "z"')                -- sigil + functional pipe stages

-- value pipelines (thru) + bitwise coexistence
do
	local out = codegen.rewrite("local p = $PATH | string.split ':'")
	check('value pipe -> thru', out:find('paper.thru', 1, true) ~= nil, out)
end
do
	local out = codegen.rewrite('local f = A | B')
	check('bitwise stays plain |', out:find('paper.thru', 1, true) == nil and out:find('A | B', 1, true) ~= nil, out)
end
valid("local n = $(printf '7') | tonumber")
valid('$(whoami) | string.upper | string.lower')
valid('ls | grep x')                 -- shell pipeline still on snail path
valid("x = a | b | c")               -- bitwise chain

-- source map (codegen.lastMap) ---------------------------------------------

do -- top-level: each statement's anchor resolves back to its own source span
	local src = 'local n = 0\nls -la'
	codegen.rewrite(src)
	local map = codegen.lastMap()
	check('map has 2 anchors', #map == 2, '#map=' .. #map)
	if #map == 2 then
		check('anchor1 resolves', src:sub(map[1].srcPos, map[1].srcPos + map[1].len - 1) == 'local n = 0',
			src:sub(map[1].srcPos, map[1].srcPos + map[1].len - 1))
		check('anchor2 genLine', map[2].genLine == 2, 'genLine=' .. tostring(map[2].genLine))
		check('anchor2 resolves', src:sub(map[2].srcPos, map[2].srcPos + map[2].len - 1) == 'ls -la',
			src:sub(map[2].srcPos, map[2].srcPos + map[2].len - 1))
	end
end

do -- nested function body: inner statements get anchors at absolute (not
   -- fragment-relative) source positions
	local src = 'local f = function()\n  ls\n  git status\nend'
	codegen.rewrite(src)
	local map = codegen.lastMap()
	local found = false
	for _, a in ipairs(map) do
		if src:sub(a.srcPos, a.srcPos + a.len - 1) == 'git status' then found = true end
	end
	check('nested anchor resolves to absolute source text', found)
end

do -- regression: emittedLine used to freeze across an entire nested-function
   -- render, so every inner statement's anchor got the SAME genLine instead
   -- of each advancing by one. Assert they're distinct and increasing.
	local src = 'local f = function()\n  a()\n  b()\n  c()\nend'
	codegen.rewrite(src)
	local map = codegen.lastMap()
	local lines = {}
	for _, a in ipairs(map) do
		local text = src:sub(a.srcPos, a.srcPos + a.len - 1)
		if text == 'a()' or text == 'b()' or text == 'c()' then lines[text] = a.genLine end
	end
	local detail = string.format('a=%s b=%s c=%s', tostring(lines['a()']), tostring(lines['b()']), tostring(lines['c()']))
	check('nested anchors found', lines['a()'] and lines['b()'] and lines['c()'], detail)
	if lines['a()'] and lines['b()'] and lines['c()'] then
		check('nested anchors strictly increase', lines['a()'] < lines['b()'] and lines['b()'] < lines['c()'], detail)
	end
end

-- diagnostics aggregation (codegen.takeDiagnostics) --------------------------

do -- clean input produces no diagnostics
	codegen.rewrite('ls -la')
	check('clean input no diags', #codegen.takeDiagnostics() == 0)
end

do -- the starter warning: an empty interpolation sigil, exact position
	codegen.rewrite('ls @()')
	local diags = codegen.takeDiagnostics()
	local d
	for _, x in ipairs(diags) do if x.code == 'empty-sigil' then d = x end end
	check('empty sigil warning found', d ~= nil)
	if d then
		check('empty sigil is a warning', d.severity == 'warning', tostring(d.severity))
		check('empty sigil position', d.labels[1].line == 1 and d.labels[1].col == 4,
			'line=' .. tostring(d.labels[1].line) .. ' col=' .. tostring(d.labels[1].col))
	end
end

do -- lexer diagnostics (lexical errors) surface through codegen too
	codegen.rewrite('local b = $(git status')
	local diags = codegen.takeDiagnostics()
	local d
	for _, x in ipairs(diags) do if x.code == 'unterminated' then d = x end end
	check('lexer diag surfaces via codegen', d ~= nil)
end

do -- classifier diagnostics (structural errors) surface through codegen too
	codegen.rewrite('end')
	local diags = codegen.takeDiagnostics()
	local d
	for _, x in ipairs(diags) do if x.code == 'stray-terminator' then d = x end end
	check('classifier diag surfaces via codegen', d ~= nil)
end

do -- nested-body diagnostics resolve to an ABSOLUTE position against the top-
   -- level source, not the nested fragment's own (re-tokenized-from-1) numbering
	local src = 'local f = function()\n  ls @()\nend'
	codegen.rewrite(src)
	local diags = codegen.takeDiagnostics()
	local d
	for _, x in ipairs(diags) do if x.code == 'empty-sigil' then d = x end end
	check('nested empty sigil warning found', d ~= nil)
	if d then
		check('nested empty sigil absolute position', d.labels[1].line == 2 and d.labels[1].col == 6,
			'line=' .. tostring(d.labels[1].line) .. ' col=' .. tostring(d.labels[1].col))
	end
end

do -- takeDiagnostics returns AND clears - it must not leak into the next rewrite
	codegen.rewrite('ls @()')
	codegen.takeDiagnostics()
	codegen.rewrite('ls -la')
	check('takeDiagnostics clears between rewrites', #codegen.takeDiagnostics() == 0)
end

-- summary -----------------------------------------------------------------

if failures == 0 then
	print(('ok - %d assertions passed'):format(total))
	os.exit(0)
end
io.stderr:write(('%d/%d assertions failed\n'):format(failures, total))
os.exit(1)
