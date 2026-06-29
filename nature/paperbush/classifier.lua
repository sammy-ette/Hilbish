local lexer = require 'nature.paperbush.lexer'
local k = lexer.kinds

local classifier = {}

local opens  = { [k.LPAREN] = true, [k.LBRACE] = true, [k.LBRACKET] = true }
local closes = { [k.RPAREN] = true, [k.RBRACE] = true, [k.RBRACKET] = true }

-- after a leading NAME, any of these means the statement is Lua (rule 2)
local luaSignal = {
	[k.ASSIGN] = true, [k.COMMA] = true, [k.DOT] = true, [k.COLON] = true,
	[k.LPAREN] = true, [k.LBRACE] = true, [k.LBRACKET] = true,
}

local parseBlock, parseStatement, parseIf, parseLoop, parseDo, parseRepeat, parseFunction

-- set by any block parser that hits EOF instead of its expected closer; read
-- back by classifier.parse so validate() can prompt for continuation instead
-- of silently compiling a half-finished block.
local sawIncomplete = false

-- structural diagnostics collected during the current classifier.parse call.
-- Unlike sawIncomplete (input that's merely unfinished), these are genuinely
-- wrong: mismatched brackets, stray terminators, missing then/do, malformed
-- pipelines. Read back as parse's third return.
local diags = {}

local closerChar = { [k.LPAREN] = ')', [k.LBRACE] = '}', [k.LBRACKET] = ']' }
local openerChar = { [k.RPAREN] = '(', [k.RBRACE] = '{', [k.RBRACKET] = '[' }
local pairOpener = { [k.RPAREN] = k.LPAREN, [k.RBRACE] = k.LBRACE, [k.RBRACKET] = k.LBRACKET }

-- checkBrackets validates () [] {} nesting across the whole token stream with
-- a typed stack: a closer that doesn't match the innermost opener is a real
-- error (wrong structure), not a continuation case. Leftover openers at EOF
-- are returned to the caller, who marks that incomplete instead.
local function checkBrackets(toks)
	local stack = {}
	for _, t in ipairs(toks) do
		if opens[t.kind] then
			stack[#stack + 1] = t
		elseif closes[t.kind] then
			local top = stack[#stack]
			if not top then
				diags[#diags + 1] = {
					severity = 'error', code = 'unbalanced',
					message = string.format("no '%s' to close", openerChar[t.kind]),
					labels = { { line = t.line, col = t.col, pos = t.pos, len = 1,
						text = string.format("this '%s' has no matching '%s'", t.value, openerChar[t.kind]), primary = true } },
					help = string.format("remove it, or add the opening '%s' earlier", openerChar[t.kind]),
				}
			elseif pairOpener[t.kind] ~= top.kind then
				diags[#diags + 1] = {
					severity = 'error', code = 'unbalanced',
					message = string.format("expected '%s', found '%s'", closerChar[top.kind], t.value),
					labels = {
						{ line = top.line, col = top.col, pos = top.pos, len = 1, text = 'opened here' },
						{ line = t.line, col = t.col, pos = t.pos, len = 1, text = string.format("expected '%s'", closerChar[top.kind]), primary = true },
					},
					help = string.format("close '%s' with '%s' instead", openerChar[top.kind], closerChar[top.kind]),
				}
				stack[#stack] = nil -- resync so it doesn't cascade
			else
				stack[#stack] = nil
			end
		end
	end
	return stack
end

local function hasTopLevelPipe(toks)
	local depth = 0
	for _, t in ipairs(toks) do
		if opens[t.kind] then depth = depth + 1
		elseif closes[t.kind] then depth = depth - 1
		elseif t.kind == k.PIPE and depth == 0 then return true end
	end
	return false
end

-- The pipeline's first stage decides its mode. A value/sigil head ($VAR, $(),
-- @(), !(), $[], number, string, paren) => an all-Lua value pipeline run through
-- `paper.thru`. A NAME head (a command or a dotted ref) => the snail pipeline
-- path, which feeds text between stages.
local valueHead = {
	[k.ENV] = true, [k.CAPTURE] = true, [k.RESULT] = true, [k.RUN] = true,
	[k.EVAL] = true, [k.NUMBER] = true, [k.STRING] = true, [k.LONGSTRING] = true,
	[k.LPAREN] = true,
}
local function headIsValue(toks, start)
	for idx = start, #toks do
		local t = toks[idx]
		if t.kind == k.EOF then return false end
		if t.kind ~= k.NEWLINE and t.kind ~= k.SEMI then
			return valueHead[t.kind] == true
		end
	end
	return false
end

local function assignRegion(toks)
	local depth, block, pendingDo = 0, 0, false
	for idx, t in ipairs(toks) do
		if opens[t.kind] then depth = depth + 1
		elseif closes[t.kind] then depth = depth - 1
		elseif t.kind == k.ASSIGN and depth == 0 and block == 0 then return idx + 1
		elseif t.kind == k.KEYWORD then
			local v = t.value
			if v == 'function' or v == 'if' or v == 'repeat' then block = block + 1
			elseif v == 'for' or v == 'while' then block = block + 1; pendingDo = true
			elseif v == 'do' then if pendingDo then pendingDo = false else block = block + 1 end
			elseif v == 'end' or v == 'until' then block = block - 1 end
		end
	end
	return 1
end

-- a bare path command (`./foo`, `../foo`, `/foo`, `~/foo`) starts with `.`/`/`/`~`
-- tokens before the executable's NAME. Those lex as DOT/OP (never NAME), so the
-- `first.kind ~= NAME` check below would otherwise wrongly file these as Lua --
-- a leading dot or slash is never valid at the start of a Lua statement anyway.
local pathChars = { ['.'] = true, ['/'] = true, ['~'] = true }
local function isPathChars(s)
	for c in s:gmatch('.') do
		if not pathChars[c] then return false end
	end
	return true
end
local function pathCommandHead(toks)
	local i, sawSlash = 1, false
	while toks[i] do
		local t = toks[i]
		if t.kind == k.DOT then
			i = i + 1
		elseif t.kind == k.OP and isPathChars(t.value) then
			if t.value:find('/', 1, true) then sawSlash = true end
			i = i + 1
		else
			break
		end
	end
	if i == 1 or not sawSlash then return false end
	return toks[i] ~= nil and toks[i].kind == k.NAME
end

local function classifyOne(toks, locals)
	local first = toks[1]
	if not first then return 'lua' end
	if first.kind == k.KEYWORD then return 'lua' end

	local second = toks[2]
	if second and (second.kind == k.ASSIGN or second.kind == k.COMMA) then return 'lua' end

	if hasTopLevelPipe(toks) then
		return headIsValue(toks, 1) and 'lua' or 'pipeline'
	end

	if first.kind ~= k.NAME then
		if pathCommandHead(toks) then return 'command' end
		return 'lua'
	end
	if not second then return 'command' end
	if luaSignal[second.kind] then return 'lua' end

	if second.kind == k.STRING or second.kind == k.LONGSTRING then
		if #toks > 2 then return 'command' end
		if locals[first.value] then return 'lua' end
		return 'dispatch'
	end

	return 'command'
end

local function declaredLocals(toks)
	local names = {}
	local first = toks[1]
	if not (first and first.kind == k.KEYWORD and first.value == 'local') then return names end
	if toks[2] and toks[2].kind == k.KEYWORD and toks[2].value == 'function' then
		if toks[3] and toks[3].kind == k.NAME then names[#names + 1] = toks[3].value end
	else
		local i = 2
		while toks[i] and toks[i].kind == k.NAME do
			names[#names + 1] = toks[i].value
			i = i + 1
			if toks[i] and toks[i].kind == k.COMMA then i = i + 1 else break end
		end
	end
	return names
end

local function gatherSimple(toks, i, src)
	local startPos = toks[i].pos
	local cur, bracket, block, pendingDo = {}, 0, 0, false
	while toks[i] and toks[i].kind ~= k.EOF do
		local t = toks[i]
		if t.kind == k.SEMI or t.kind == k.NEWLINE then
			if bracket == 0 and block == 0 then break end
		elseif t.kind == k.KEYWORD then
			local v = t.value
			if v == 'end' or v == 'until' then
				if bracket == 0 and block == 0 then break end
				block = block - 1
			elseif v == 'else' or v == 'elseif' then
				if bracket == 0 and block == 0 then break end
			elseif v == 'function' or v == 'if' or v == 'repeat' then
				block = block + 1
			elseif v == 'for' or v == 'while' then
				block = block + 1
				pendingDo = true
			elseif v == 'do' then
				if pendingDo then pendingDo = false else block = block + 1 end
			end
		elseif opens[t.kind] then
			bracket = bracket + 1
		elseif closes[t.kind] then
			bracket = bracket - 1
		end
		cur[#cur + 1] = t
		i = i + 1
	end
	-- hit EOF while still inside an inline bracket/block (e.g. `print(`,
	-- `local f = function() ls`) => the line genuinely isn't finished yet.
	if (not toks[i] or toks[i].kind == k.EOF) and (bracket > 0 or block > 0) then
		sawIncomplete = true
	end
	local endPos = (toks[i] and toks[i].pos - 1) or #src
	return cur, i, startPos, endPos
end

-- gatherExpr slices a raw Lua span from i to the next depth-0 separator/EOF
-- (used for `until <expr>`). Keywords are not terminators here.
local function gatherExpr(toks, i, src)
	local startPos = toks[i].pos
	local depth = 0
	while toks[i] and toks[i].kind ~= k.EOF do
		local t = toks[i]
		if depth == 0 and (t.kind == k.SEMI or t.kind == k.NEWLINE) then break end
		if opens[t.kind] then depth = depth + 1
		elseif closes[t.kind] then depth = depth - 1 end
		i = i + 1
	end
	if depth > 0 then sawIncomplete = true end
	return src:sub(startPos, (toks[i] and toks[i].pos - 1) or #src), i, startPos
end
local function spanThrough(toks, i, src, stop)
	local startKw = toks[i]
	local startPos = startKw.pos
	local bracket, block, pendingDo = 0, 0, false
	local j = i + 1 -- skip the opener keyword (if / elseif / for / while)
	local closed, missing = false, false
	while toks[j] and toks[j].kind ~= k.EOF do
		local t = toks[j]
		if opens[t.kind] then
			bracket = bracket + 1
		elseif closes[t.kind] then
			bracket = bracket - 1
		elseif t.kind == k.KEYWORD then
			local v = t.value
			if v == 'do' then
				if pendingDo then pendingDo = false
				elseif bracket == 0 and block == 0 and stop[v] then closed = true; j = j + 1; break
				else block = block + 1 end
			elseif stop[v] and bracket == 0 and block == 0 then
				closed = true; j = j + 1; break
			elseif v == 'function' or v == 'if' or v == 'repeat' then
				block = block + 1
			elseif v == 'for' or v == 'while' then
				block = block + 1; pendingDo = true
			elseif v == 'end' or v == 'until' then
				-- this end/until isn't ours to consume: it belongs to an
				-- enclosing block, which means our header never got its
				-- then/do. That's wrong, not unfinished - stop here and let
				-- the caller (parseBlock) see this token fresh.
				if bracket == 0 and block == 0 then missing = true; break end
				block = block - 1
			end
		end
		j = j + 1
	end
	if missing then
		local expected = next(stop)
		diags[#diags + 1] = {
			severity = 'error', code = 'missing-keyword',
			message = string.format("missing '%s' after '%s'", expected, startKw.value),
			labels = { { line = startKw.line, col = startKw.col, pos = startKw.pos, len = #startKw.value,
				text = string.format("this '%s' has no '%s'", startKw.value, expected), primary = true } },
			help = string.format("add '%s' after the condition", expected),
		}
	elseif not closed then
		sawIncomplete = true
	end
	return src:sub(startPos, (toks[j] and toks[j].pos - 1) or #src), j, startPos
end

local function spanThroughParens(toks, i, src)
	local startPos = toks[i].pos
	local depth, seen, closed = 0, false, false
	while toks[i] and toks[i].kind ~= k.EOF do
		local t = toks[i]
		if opens[t.kind] then depth = depth + 1; seen = true
		elseif closes[t.kind] then
			depth = depth - 1
			if seen and depth == 0 then closed = true; i = i + 1; break end
		end
		i = i + 1
	end
	if not closed then sawIncomplete = true end
	return src:sub(startPos, (toks[i] and toks[i].pos - 1) or #src), i, startPos
end

local function splitPipeline(toks, src, stmtEnd, locals, atEOF)
	local stages, seg, depth = {}, {}, 0
	local sawStage, lastPipe = false, nil
	local function flush(endPos)
		if #seg == 0 then return end
		local first = seg[1]
		if #seg == 1 and first.kind == k.EVAL then
			stages[#stages + 1] = { kind = 'lua', expr = first.value, pos = first.pos + 2 }
		elseif first.kind == k.KEYWORD and first.value == 'function' then
			stages[#stages + 1] = { kind = 'lua', expr = src:sub(first.pos, endPos), pos = first.pos }
		elseif first.kind == k.NAME and ((seg[2] and (seg[2].kind == k.DOT or seg[2].kind == k.COLON)) or locals[first.value]) then
			stages[#stages + 1] = {kind = 'func', src = src:sub(first.pos, endPos), pos = first.pos}
		else
			stages[#stages + 1] = { kind = 'shell', src = src:sub(first.pos, endPos), pos = first.pos }
		end
		sawStage = true
		seg = {}
	end
	for _, t in ipairs(toks) do
		if opens[t.kind] then depth = depth + 1; seg[#seg + 1] = t
		elseif closes[t.kind] then depth = depth - 1; seg[#seg + 1] = t
		elseif t.kind == k.PIPE and depth == 0 then
			if #seg == 0 then
				if not sawStage then
					diags[#diags + 1] = {
						severity = 'error', code = 'empty-pipeline-stage',
						message = "pipeline starts with '|'",
						labels = { { line = t.line, col = t.col, pos = t.pos, len = 1,
							text = 'a pipeline must start with a command or value, not |', primary = true } },
						help = 'remove the leading |, or put a command/value before it',
					}
				else
					diags[#diags + 1] = {
						severity = 'error', code = 'empty-pipeline-stage',
						message = 'empty stage between pipes',
						labels = { { line = t.line, col = t.col, pos = t.pos, len = 1, text = 'nothing between the pipes here', primary = true } },
						help = 'remove one of the |, or add a command/value between them',
					}
				end
			else
				flush(t.pos - 1)
			end
			lastPipe = t
		else seg[#seg + 1] = t end
	end
	if #seg == 0 and lastPipe then
		if atEOF then
			sawIncomplete = true
		else
			diags[#diags + 1] = {
				severity = 'error', code = 'empty-pipeline-stage',
				message = "missing a command or value after '|'",
				labels = { { line = lastPipe.line, col = lastPipe.col, pos = lastPipe.pos, len = 1, text = 'nothing follows this', primary = true } },
				help = 'remove the trailing |, or add a command/value after it',
			}
		end
	else
		flush(stmtEnd)
	end
	return stages
end

-- add the loop variables of a `for` header to the locals set
local function addForVars(toks, i, locals)
	local j = i + 1
	while toks[j] and toks[j].kind == k.NAME do
		locals[toks[j].value] = true
		j = j + 1
		if toks[j] and toks[j].kind == k.COMMA then j = j + 1 else break end
	end
end

-- closeBlock appends the `end` closer part for for/while/do/function blocks.
-- If EOF was reached instead of a real `end` keyword, the block is genuinely
-- unfinished (more input is needed), not a syntax error to surface immediately.
local function closeBlock(toks, i, parts, src)
	local atEnd = toks[i] and toks[i].kind == k.KEYWORD and toks[i].value == 'end'
	if not atEnd then sawIncomplete = true end
	parts[#parts + 1] = { span = 'end', pos = (toks[i] and toks[i].pos) or #src }
	if toks[i] then i = i + 1 end
	return i
end

function parseIf(toks, i, src, locals)
	local parts = {}
	local hdr, hpos
	hdr, i, hpos = spanThrough(toks, i, src, { ['then'] = true })
	parts[#parts + 1] = { span = hdr, pos = hpos }
	while true do
		local body
		body, i = parseBlock(toks, i, src, locals, { ['elseif'] = true, ['else'] = true, ['end'] = true })
		parts[#parts + 1] = { body = body }
		local kw = toks[i] and toks[i].kind == k.KEYWORD and toks[i].value
		if kw == 'elseif' then
			hdr, i, hpos = spanThrough(toks, i, src, { ['then'] = true })
			parts[#parts + 1] = { span = hdr, pos = hpos }
		elseif kw == 'else' then
			parts[#parts + 1] = { span = 'else', pos = toks[i].pos }
			i = i + 1
		else -- end (or EOF)
			if kw ~= 'end' then sawIncomplete = true end
			parts[#parts + 1] = { span = 'end', pos = (toks[i] and toks[i].pos) or #src }
			if toks[i] then i = i + 1 end
			break
		end
	end
	return { type = 'block', parts = parts }, i
end

function parseLoop(toks, i, src, locals)
	if toks[i].value == 'for' then addForVars(toks, i, locals) end
	local parts = {}
	local hdr, hpos
	hdr, i, hpos = spanThrough(toks, i, src, { ['do'] = true })
	parts[#parts + 1] = { span = hdr, pos = hpos }
	local body
	body, i = parseBlock(toks, i, src, locals, { ['end'] = true })
	parts[#parts + 1] = { body = body }
	i = closeBlock(toks, i, parts, src)
	return { type = 'block', parts = parts }, i
end

function parseDo(toks, i, src, locals)
	local parts = { { span = 'do', pos = toks[i].pos } }
	i = i + 1
	local body
	body, i = parseBlock(toks, i, src, locals, { ['end'] = true })
	parts[#parts + 1] = { body = body }
	i = closeBlock(toks, i, parts, src)
	return { type = 'block', parts = parts }, i
end

function parseRepeat(toks, i, src, locals)
	local parts = { { span = 'repeat', pos = toks[i].pos } }
	i = i + 1
	local body
	body, i = parseBlock(toks, i, src, locals, { ['until'] = true })
	parts[#parts + 1] = { body = body }
	if toks[i] and toks[i].value == 'until' then
		local span, spos
		span, i, spos = gatherExpr(toks, i, src)
		parts[#parts + 1] = { span = span, pos = spos }
	else
		sawIncomplete = true
	end
	return { type = 'block', parts = parts }, i
end

function parseFunction(toks, i, src, locals)
	-- record the function name as callable (so `f 'x'` later is a direct call)
	local nameIdx = (toks[i].value == 'local') and i + 2 or i + 1
	if toks[nameIdx] and toks[nameIdx].kind == k.NAME then locals[toks[nameIdx].value] = true end
	local parts = {}
	local hdr, hpos
	hdr, i, hpos = spanThroughParens(toks, i, src)
	parts[#parts + 1] = { span = hdr, pos = hpos }
	local body
	body, i = parseBlock(toks, i, src, locals, { ['end'] = true })
	parts[#parts + 1] = { body = body }
	i = closeBlock(toks, i, parts, src)
	return { type = 'block', parts = parts }, i
end

function parseStatement(toks, i, src, locals)
	local t = toks[i]
	if t.kind == k.KEYWORD then
		local kw = t.value
		local block, ni
		if kw == 'function' then block, ni = parseFunction(toks, i, src, locals) end
		if kw == 'local' and toks[i + 1] and toks[i + 1].kind == k.KEYWORD and toks[i + 1].value == 'function' then
			block, ni = parseFunction(toks, i, src, locals)
		end
		if kw == 'if' then block, ni = parseIf(toks, i, src, locals) end
		if kw == 'for' or kw == 'while' then block, ni = parseLoop(toks, i, src, locals) end
		if kw == 'do' then block, ni = parseDo(toks, i, src, locals) end
		if kw == 'repeat' then block, ni = parseRepeat(toks, i, src, locals) end
		if block then
			block.startPos = t.pos
			block.endPos = (toks[ni] and toks[ni].pos - 1) or #src
			return block, ni
		end
		-- other keyword-led statements (local var, return, break, goto) fall through
	end

	local tokens, ni, startPos, endPos = gatherSimple(toks, i, src)
	local stmt = {
		startPos = startPos, endPos = endPos,
		src = src:sub(startPos, endPos), tokens = tokens,
	}
	stmt.type = classifyOne(tokens, locals)
	if stmt.type == 'pipeline' then
		local atEOF = (not toks[ni]) or toks[ni].kind == k.EOF
		stmt.stages = splitPipeline(tokens, src, endPos, locals, atEOF)
	end
	if stmt.type == 'dispatch' then stmt.head = tokens[1].value end
	if stmt.type == 'lua' and hasTopLevelPipe(tokens)
		and headIsValue(tokens, assignRegion(tokens)) then
		stmt.thru = true
	end
	for _, name in ipairs(declaredLocals(tokens)) do locals[name] = true end
	return stmt, ni
end

-- block-closing keywords that, if reached here (i.e. not consumed by a stop
-- set above, meaning no enclosing block claims them), have nothing to close.
local strayKeyword = { ['end'] = true, ['else'] = true, ['elseif'] = true, ['until'] = true }

function parseBlock(toks, i, src, locals, stop)
	local stmts = {}
	while toks[i] and toks[i].kind ~= k.EOF do
		local t = toks[i]
		if t.kind == k.SEMI or t.kind == k.NEWLINE then
			i = i + 1
		elseif t.kind == k.KEYWORD and stop[t.value] then
			break
		else
			local before = i
			local stmt
			stmt, i = parseStatement(toks, i, src, locals)
			if i == before then
				if t.kind == k.KEYWORD and strayKeyword[t.value] then
					diags[#diags + 1] = {
						severity = 'error', code = 'stray-terminator',
						message = string.format("'%s' with no matching block to close", t.value),
						labels = { { line = t.line, col = t.col, pos = t.pos, len = #t.value, text = 'unexpected here', primary = true } },
						help = string.format("remove this '%s', or add the block it is meant to close", t.value),
					}
				end
				i = i + 1
			else
				stmts[#stmts + 1] = stmt
			end
		end
	end
	return stmts, i
end

function classifier.parse(src)
	sawIncomplete = false
	diags = {}
	local toks = lexer.tokenize(src)
	local leftover = checkBrackets(toks)
	if #leftover > 0 then sawIncomplete = true end
	local stmts = (parseBlock(toks, 1, src, {}, {}))
	return stmts, sawIncomplete, diags
end

return classifier
