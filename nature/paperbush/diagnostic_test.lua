-- Tests for the Paperbush diagnostic renderer.
--
-- Run under Hilbish from the repo root:
--   hilbish nature/paperbush/diagnostic_test.lua
-- Exits 0 if all assertions pass, 1 otherwise.

package.path = './?.lua;' .. package.path
local diagnostic = require('nature.paperbush.diagnostic')

local total, failures = 0, 0
local function fail(name, detail)
	failures = failures + 1
	io.stderr:write(string.format('FAIL: %s%s\n', name, detail and ('\n      ' .. detail) or ''))
end
local function check(name, cond, detail)
	total = total + 1
	if not cond then fail(name, detail) end
end

local function lines(s)
	local out = {}
	for l in (s .. '\n'):gmatch('([^\n]*)\n') do out[#out + 1] = l end
	return out
end

local function hasAnsi(s) return s:find('\27[', 1, true) ~= nil end

-- single label --------------------------------------------------------------

do
	local src = 'print) hi'
	local diag = {
		severity = 'error', code = 'unbalanced',
		message = "no '(' to close",
		labels = { { line = 1, col = 6, len = 1, text = "this ')' has no matching '('", primary = true } },
		help = "remove it, or add the opening '(' earlier",
	}
	local out = diagnostic.render(diag, src, { color = false })
	local ls = lines(out)
	check('header has severity+code+message', out:find('error%[unbalanced%]: ' .. "no '%(' to close") ~= nil, out)
	check('location rail present', out:find('paperbush:1:6', 1, true) ~= nil, out)
	check('source line shown', out:find('print%) hi') ~= nil, out)
	check('help line present', out:find('help: ' .. "remove it", 1, true) ~= nil, out)
	check('no ANSI when color=false', not hasAnsi(out), out)

	-- the caret row must be directly below the source row, with the caret
	-- under column 6 (5 chars of indent before it)
	local srcIdx
	for i, l in ipairs(ls) do if l:find('print%) hi') then srcIdx = i end end
	check('source line found', srcIdx ~= nil)
	if srcIdx then
		local caretLine = ls[srcIdx + 1]
		check('caret line has ^', caretLine ~= nil and caretLine:find('%^') ~= nil, caretLine)
		if caretLine then
			local caretCol = caretLine:find('%^')
			local srcCol = ls[srcIdx]:find('%)')
			check('caret aligns under the )', caretCol == srcCol,
				'caretCol=' .. tostring(caretCol) .. ' srcCol=' .. tostring(srcCol))
		end
	end
end

-- multi-label (opener + EOF), unordered input gets sorted by line ----------

do
	local src = 'local b = $(git status'
	local diag = {
		severity = 'error', code = 'unterminated',
		message = "unclosed '$(' capture",
		labels = {
			{ line = 1, col = 24, len = 1, text = "expected ')' before end of input" },
			{ line = 1, col = 11, len = 2, text = 'opened here', primary = true },
		},
	}
	local out = diagnostic.render(diag, src, { color = false })
	check('both labels rendered', out:find('opened here', 1, true) ~= nil
		and out:find("expected ')' before end of input", 1, true) ~= nil, out)
	check('primary uses ^, secondary uses ~',
		out:find('%^%^ opened here') ~= nil and out:find('~ expected') ~= nil, out)
	-- the source line should only be printed ONCE even though both labels are on line 1
	local srcOccurrences = 0
	for _ in out:gmatch('local b = %$%(git status') do srcOccurrences = srcOccurrences + 1 end
	check('source line printed once for same-line labels', srcOccurrences == 1, tostring(srcOccurrences))
end

-- severities ------------------------------------------------------------------

do
	local function headerFor(sev)
		return diagnostic.render({ severity = sev, message = 'x', labels = {} }, 'src', { color = false })
	end
	check('error header', headerFor('error'):find('^error:', 1, false) ~= nil or headerFor('error'):find('error: x', 1, true) ~= nil)
	check('warning header', headerFor('warning'):find('warning: x', 1, true) ~= nil)
	check('note header', headerFor('note'):find('note: x', 1, true) ~= nil)
	check('unknown severity falls back to error', headerFor('bogus'):find('error: x', 1, true) ~= nil)
end

-- color on/off + NO_COLOR ----------------------------------------------------

do
	local diag = { severity = 'error', message = 'boom',
		labels = { { line = 1, col = 1, len = 1, primary = true } } }
	local colored = diagnostic.render(diag, 'x', { color = true })
	local plain = diagnostic.render(diag, 'x', { color = false })
	check('color=true produces ANSI', hasAnsi(colored), colored)
	check('color=false produces no ANSI', not hasAnsi(plain), plain)
end

do
	local diag = { severity = 'error', message = 'boom',
		labels = { { line = 1, col = 1, len = 1, primary = true } } }
	local prev = os.getenv('NO_COLOR')
	os.setenv('NO_COLOR', '1')
	local out = diagnostic.render(diag, 'x')
	os.setenv('NO_COLOR', prev or '')
	check('NO_COLOR=1 disables color by default', not hasAnsi(out), out)
end

-- graceful degradation --------------------------------------------------------

do -- a label pointing past the end of the source degrades to message-only
	local diag = {
		severity = 'warning', message = 'out of range', help = 'still helpful',
		labels = { { line = 99, col = 1, len = 1, primary = true } },
	}
	local out = diagnostic.render(diag, 'one\ntwo', { color = false })
	check('out-of-range degrades, no crash', out ~= nil)
	check('message-only keeps the message', out:find('out of range', 1, true) ~= nil, out)
	check('message-only keeps the help', out:find('still helpful', 1, true) ~= nil, out)
	check('message-only has no location rail', out:find('┌─', 1, true) == nil, out)
end

do -- a label with a line but no column shows the source line, no caret row
	local diag = {
		severity = 'error', message = 'whole-line issue',
		labels = { { line = 1, text = 'somewhere on this line', primary = true } },
	}
	local out = diagnostic.render(diag, 'ls -la', { color = false })
	check('line-only label shows source', out:find('ls %-la') ~= nil, out)
	check('line-only label has no caret', out:find('%^') == nil, out)
end

-- tab expansion ---------------------------------------------------------------

do -- a tab before the caret position must still land the caret on the right column
	local src = '\tfoo)'
	local diag = {
		severity = 'error', message = 'x',
		labels = { { line = 1, col = 5, len = 1, primary = true } }, -- the ')'
	}
	local out = diagnostic.render(diag, src, { color = false })
	local ls = lines(out)
	local srcIdx
	for i, l in ipairs(ls) do if l:find('foo%)') then srcIdx = i end end
	check('tab-expanded source line found', srcIdx ~= nil)
	if srcIdx then
		local caretLine = ls[srcIdx + 1]
		local caretCol = caretLine and caretLine:find('%^')
		local srcCol = ls[srcIdx]:find('%)')
		check('caret aligns under ) despite the leading tab', caretCol == srcCol,
			'caretCol=' .. tostring(caretCol) .. ' srcCol=' .. tostring(srcCol))
	end
end

-- report() writes render()'s output to stderr --------------------------------

do
	local diag = { severity = 'error', message = 'reported', labels = {} }
	local captured = {}
	local realStderr = io.stderr
	io.stderr = { write = function(_, s) captured[#captured + 1] = s end }
	diagnostic.report(diag, 'src', { color = false })
	io.stderr = realStderr
	local got = table.concat(captured)
	local want = diagnostic.render(diag, 'src', { color = false })
	check('report writes render output + newline', got == want .. '\n',
		('%q vs %q'):format(got, want))
end

-- summary -----------------------------------------------------------------

if failures == 0 then
	print(string.format('ok - %d assertions passed', total))
	os.exit(0)
end

io.stderr:write(string.format('%d/%d assertions failed\n', failures, total))
os.exit(1)
