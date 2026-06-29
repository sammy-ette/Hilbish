-- Tests for the Paperbush runtime, against the REAL Hilbish runtime.
-- Run under Hilbish from the repo root:
--   hilbish nature/paperbush/runtime_test.lua
-- Exits 0 if all pass, 1 otherwise.

package.path = './?.lua;' .. package.path
local M = require('nature.paperbush.runtime')

local total, failures = 0, 0
local function fail(n, d)
	failures = failures + 1
	io.stderr:write('FAIL: ' .. n .. (d and ('\n      ' .. d) or '') .. '\n')
end
local function check(name, cond, detail)
	total = total + 1
	if not cond then fail(name, detail) end
end

-- capture: stdout, trailing whitespace stripped
check('capture echo', M.capture('echo hi') == 'hi', M.capture('echo hi'))

-- result: exit code + streams
check('result true', M.result('true').exitCode == 0)
check('result false', M.result('false').exitCode ~= 0)
do
	local r = M.result('echo out')
	check('result stdout', (r.stdout or ''):match('out') ~= nil, r.stdout)
end

-- exec: uncaptured, returns the exit code
check('exec true', M.exec('true') == 0)
check('exec false', M.exec('false') ~= 0)

-- env: $VAR lookup (pcall'd because M.env currently references an undefined `env`)
do
	local ok, home = pcall(M.env, 'HOME')
	check('env HOME', ok and home ~= '', ok and 'empty' or tostring(home))
	local ok2, none = pcall(M.env, 'PAPERBUSH_DEFINITELY_UNSET')
	check('env unset', ok2 and none == '', ok2 and ('got ' .. tostring(none)) or tostring(none))
end

-- quote: single shell word, escaped quotes, array splice
check('quote plain', M.quote('hi') == "'hi'", M.quote('hi'))
check('quote embedded', M.quote("a'b") == "'a'\\''b'", M.quote("a'b"))
check('quote array', M.quote({ '-l', 'my dir' }) == "'-l' 'my dir'", M.quote({ '-l', 'my dir' }))

-- dispatch: callable global is called; otherwise the command runs
do
	_G.pb_dispatch_probe = function(a) return 'GOT:' .. a end
	check('dispatch callable', M.dispatch('pb_dispatch_probe', { 'x' }, 'pb_dispatch_probe x') == 'GOT:x')
	check('dispatch fallback', M.dispatch('pb_no_such_cmd_xyz', {}, 'true') == 0)
	_G.pb_dispatch_probe = nil
end

if failures == 0 then
	print(('ok - %d assertions passed'):format(total))
	os.exit(0)
end
io.stderr:write(('%d/%d assertions failed\n'):format(failures, total))
os.exit(1)
