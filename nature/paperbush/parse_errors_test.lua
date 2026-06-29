-- Tests for Paperbush's native syntax diagnostics: the lexer's lexical errors
-- (lexer.diagnose) and the classifier's structural errors (classifier.parse's
-- third return). The key property under test, beyond exact positions, is the
-- incomplete/error split: input that's merely unfinished (more could still be
-- typed) must stay `incomplete` with no diagnostic, while input that's
-- genuinely wrong must produce a diagnostic and NOT be flagged incomplete.
--
-- Run under Hilbish from the repo root:
--   hilbish nature/paperbush/parse_errors_test.lua
-- Exits 0 if all assertions pass, 1 otherwise.

package.path = './?.lua;' .. package.path
local lexer = require('nature.paperbush.lexer')
local classifier = require('nature.paperbush.classifier')

local total, failures = 0, 0
local function fail(name, detail)
	failures = failures + 1
	io.stderr:write(string.format('FAIL: %s%s\n', name, detail and ('\n      ' .. detail) or ''))
end
local function check(name, cond, detail)
	total = total + 1
	if not cond then fail(name, detail) end
end

local function lexDiags(src)
	return lexer.diagnose(lexer.tokenize(src), src)
end

local function findDiag(diags, code)
	for _, d in ipairs(diags) do if d.code == code then return d end end
	return nil
end

-- diagAt asserts some diagnostic with `code` exists in `diags` whose message
-- contains `msgPart`, returning it for further (e.g. position) checks.
local function diagAt(name, diags, code, msgPart)
	total = total + 1
	local d = findDiag(diags, code)
	if not d then
		fail(name, 'no diagnostic with code ' .. code .. '; got ' .. #diags .. ' diag(s)')
		return nil
	end
	if msgPart and not d.message:find(msgPart, 1, true) then
		fail(name, 'message ' .. ('%q'):format(d.message) .. ' missing ' .. ('%q'):format(msgPart))
		return nil
	end
	return d
end

local function noDiags(name, diags)
	total = total + 1
	if #diags ~= 0 then
		local msgs = {}
		for _, d in ipairs(diags) do msgs[#msgs + 1] = d.code .. ': ' .. d.message end
		fail(name, 'expected no diagnostics, got: ' .. table.concat(msgs, ' | '))
	end
end

local function labelAt(name, label, line, col)
	total = total + 1
	if not (label and label.line == line and label.col == col) then
		fail(name, string.format('expected %d:%d, got %s:%s',
			line, col, tostring(label and label.line), tostring(label and label.col)))
	end
end

-- lexer: lexical diagnostics ----------------------------------------------

do -- illegal byte, exact position
	local diags = lexDiags('echo `bad`')
	local d = diagAt('illegal char', diags, 'illegal-char')
	if d then labelAt('illegal char position', d.labels[1], 1, 6) end
end

do -- unterminated short string: opener + EOF labels
	local diags = lexDiags('echo "abc')
	local d = diagAt('unterminated string', diags, 'unterminated', 'unterminated string')
	if d then
		check('unterminated string has 2 labels', #d.labels == 2, '#labels=' .. #d.labels)
		labelAt('unterminated string opener', d.labels[1], 1, 6)
	end
end

do -- unclosed $( capture: opener + EOF labels
	local diags = lexDiags('local b = $(git status')
	local d = diagAt('unclosed capture', diags, 'unterminated', 'capture')
	if d then
		labelAt('unclosed capture opener', d.labels[1], 1, 11)
		check('unclosed capture opener len', d.labels[1].len == 2, 'len=' .. tostring(d.labels[1].len))
	end
end

do -- unclosed long string
	local diags = lexDiags('x = [[abc')
	diagAt('unclosed long string', diags, 'unterminated', 'long string')
end

do -- unclosed ${ env ref
	local diags = lexDiags('echo ${HOME')
	diagAt('unclosed ${', diags, 'unterminated', "'${'")
end

do -- malformed numbers
	diagAt('malformed 1..2', lexDiags('x = 1..2'), 'malformed-number')
	diagAt('malformed 0x alone', lexDiags('x = 0x'), 'malformed-number')
	diagAt('malformed 1e', lexDiags('x = 1e'), 'malformed-number')
end

do -- these must NOT be flagged malformed - real, valid Lua number forms
	for _, src in ipairs({ 'x = 0x1.8p3', 'x = .5', 'x = 5.', 'x = 1e10', 'x = 0xFF', 'x = 3.14' }) do
		check('valid number ' .. ('%q'):format(src), findDiag(lexDiags(src), 'malformed-number') == nil)
	end
end

do -- closed/complete constructs produce no lexical diagnostics
	noDiags('closed capture clean', lexDiags('local b = $(git status)'))
	noDiags('closed string clean', lexDiags('echo "all good"'))
end

-- classifier: structural diagnostics --------------------------------------

do -- closer with nothing open
	local _, incomplete, diags = classifier.parse('print) hi')
	local d = diagAt('no opener', diags, 'unbalanced', "no '(' to close")
	if d then labelAt('no opener position', d.labels[1], 1, 6) end
	check('no opener not incomplete', incomplete == false)
end

do -- mismatched bracket types
	local _, _, diags = classifier.parse('print(1]')
	diagAt('mismatched bracket', diags, 'unbalanced', "expected ')', found ']'")
end

do -- stray block terminators, alone, must be flagged (not silently dropped)
	for _, kw in ipairs({ 'end', 'else', 'until', 'elseif' }) do
		local _, _, diags = classifier.parse(kw)
		diagAt('stray ' .. kw, diags, 'stray-terminator', kw)
	end
end

do -- missing then/do: a real error, distinct from incomplete
	local _, incomplete, diags = classifier.parse('if x end')
	diagAt('missing then', diags, 'missing-keyword', "missing 'then' after 'if'")
	check('missing then not incomplete', incomplete == false, tostring(incomplete))
end

do
	local _, incomplete, diags = classifier.parse('for i = 1, 3 ls end')
	diagAt('missing do', diags, 'missing-keyword', "missing 'do' after 'for'")
	check('missing do not incomplete', incomplete == false, tostring(incomplete))
end

-- the critical non-regression: input that's merely unfinished stays
-- `incomplete` with NO diagnostic, even though it now goes through real
-- structural validation.
do
	local _, incomplete, diags = classifier.parse('if x then')
	check('if-then incomplete', incomplete == true)
	noDiags('if-then no diags', diags)
end

do
	local _, incomplete, diags = classifier.parse('for i = 1, 3 do')
	check('for-do incomplete', incomplete == true)
	noDiags('for-do no diags', diags)
end

do
	local _, incomplete, diags = classifier.parse('print(1')
	check('unclosed paren incomplete', incomplete == true)
	noDiags('unclosed paren no diags', diags)
end

-- pipeline shape ------------------------------------------------------------

do -- leading pipe
	local _, _, diags = classifier.parse('| grep x')
	diagAt('leading pipe', diags, 'empty-pipeline-stage', "starts with '|'")
end

do -- empty stage between two pipes
	local _, _, diags = classifier.parse('ls | | grep x')
	diagAt('empty stage', diags, 'empty-pipeline-stage', 'empty stage between pipes')
end

do -- trailing pipe at true EOF stays incomplete, no error
	local _, incomplete, diags = classifier.parse('ls |')
	check('trailing pipe at eof incomplete', incomplete == true)
	noDiags('trailing pipe at eof no diags', diags)
end

do -- trailing pipe NOT at EOF (more follows) is a real error
	local _, incomplete, diags = classifier.parse('ls |; grep x')
	diagAt('trailing pipe mid-input', diags, 'empty-pipeline-stage', "missing a command or value after '|'")
	check('trailing pipe mid-input not incomplete', incomplete == false)
end

do -- well-formed pipelines and nested blocks stay clean
	local _, incomplete, diags = classifier.parse('ls | grep x')
	check('clean pipeline not incomplete', incomplete == false)
	noDiags('clean pipeline no diags', diags)
end

do
	local _, incomplete, diags = classifier.parse('if x then for i = 1, 2 do ls end end')
	check('nested blocks not incomplete', incomplete == false)
	noDiags('nested blocks no diags', diags)
end

-- fuzz regression: the old no-hang set must still terminate, and now some of
-- them should actually surface a diagnostic instead of vanishing silently.
do
	for _, s in ipairs({ 'end', 'else', 'until', 'elseif', 'end end then' }) do
		local ok = pcall(classifier.parse, s)
		check('no-hang ' .. ('%q'):format(s), ok)
	end
	local _, _, diags = classifier.parse('end end then')
	check('end end then surfaces diags', #diags > 0)
end

-- summary -------------------------------------------------------------------

if failures == 0 then
	print(string.format('ok - %d assertions passed', total))
	os.exit(0)
end

io.stderr:write(string.format('%d/%d assertions failed\n', failures, total))
os.exit(1)
