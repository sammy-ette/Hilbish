local lexer = require 'nature.paperbush.lexer'
local classifier = require 'nature.paperbush.classifier'
local k = lexer.kinds

local M = {}

local function fmt(s) return string.format('%q', s) end

local operators = {
	[k.ENV] = true, [k.CAPTURE] = true, [k.RESULT] = true, [k.RUN] = true, [k.EVAL] = true,
}

local opens  = { [k.LPAREN] = true, [k.LBRACE] = true, [k.LBRACKET] = true }
local closes = { [k.RPAREN] = true, [k.RBRACE] = true, [k.RBRACKET] = true }

local function opSourceLen(text, t)
	if t.kind == k.ENV then
		if text:sub(t.pos + 1, t.pos + 1) == '{' then return 3 + #t.value end
		return 1 + #t.value
	end
	return 3 + #t.value
end

local rewriteLua, rewriteStatements

-- lineFinder/emittedLine keep the generated chunk's line numbers aligned with
-- the original input's, so load()/runtime error positions point at a line
-- the user actually typed instead of an arbitrary line in the rewritten code.
-- Reset per top-level M.rewrite call; shared as module upvalues since the
-- whole render is synchronous (no concurrent rewrites).
local lineFinder, emittedLine

-- anchors is the real generated-line -> source-span map: one entry per
-- statement (and per block header span), `{ genLine, srcPos, len }`. Unlike
-- the lineFinder/emittedLine cosmetic alignment above (which only tries to
-- keep the *common* case needing no remapping), this is built to be exactly
-- correct even through nested fragments, so luaerr.lua can always resolve a
-- load()/pcall() error's generated line back to the real source span that
-- produced it. Reset per top-level M.rewrite call.
local anchors

local function countNewlines(s)
	local _, n = s:gsub('\n', '')
	return n
end

-- recordAnchor takes a SOURCE position (already resolved to absolute via the
-- caller's `base`) and the current generated line (emittedLine, which is now
-- always accurate - see the nested-function-body fix in rewriteLua).
local function recordAnchor(srcPos, len)
	if not srcPos then return end
	anchors[#anchors + 1] = { genLine = emittedLine, srcPos = srcPos, len = math.max(len or 1, 1) }
end

--- Returns the source map built by the most recent M.rewrite/M.chunk call.
--- @return table list of { genLine, srcPos, len }
function M.lastMap()
	return anchors
end

-- diagnostics aggregates every front-end diagnostic (lexer + classifier +
-- codegen's own empty-sigil warning) found while rendering the current
-- M.rewrite call, across the top-level source and every nested fragment.
-- topSrc is the original input the user typed; every label's line/col is
-- resolved against it (never a fragment's own, possibly-relative, numbering).
-- Both reset per top-level M.rewrite call.
local diagnostics, topSrc

local function lineColAt(src, pos)
	local line, lastNL = 1, 0
	for p in src:gmatch('()\n') do
		if p >= pos then break end
		lastNL = p
		line = line + 1
	end
	return line, pos - lastNL
end

-- shiftDiag re-resolves a diagnostic's labels against topSrc: a label's `pos`
-- is relative to whatever fragment produced it (the top-level source when
-- base==0, otherwise a nested body re-tokenized from position 1), so its
-- line/col are only meaningful there. base + pos is always absolute; convert
-- that back to a real line/col via topSrc so every diagnostic looks the same
-- regardless of which fragment it came from.
local function shiftDiag(diag, base)
	local labels = {}
	for _, l in ipairs(diag.labels) do
		local nl = { len = l.len, text = l.text, primary = l.primary }
		if l.pos then
			nl.pos = base + l.pos
			nl.line, nl.col = lineColAt(topSrc, nl.pos)
		else
			nl.line, nl.col = l.line, l.col
		end
		labels[#labels + 1] = nl
	end
	return { severity = diag.severity, code = diag.code, message = diag.message, labels = labels, help = diag.help }
end

-- an interpolation sigil whose body is empty/whitespace-only is almost always
-- a typo (`ls @()`, `local x = $()`), not intentional - worth a warning even
-- though it's perfectly legal Lua/paperbush once rewritten.
local sigilOpener = { [k.CAPTURE] = '$(', [k.RESULT] = '!(', [k.EVAL] = '@(', [k.RUN] = '$[' }
local sigilCloser = { [k.CAPTURE] = ')', [k.RESULT] = ')', [k.EVAL] = ')', [k.RUN] = ']' }
local function emptySigilWarnings(toks, base)
	local warnings = {}
	for _, t in ipairs(toks) do
		local opener, closer, braced = sigilOpener[t.kind], sigilCloser[t.kind], false
		if not opener and t.kind == k.ENV and t.opener == '${' then
			opener, closer, braced = '${', '}', true
		end
		if opener and t.value:match('^%s*$') then
			local pos = base + t.pos
			local line, col = lineColAt(topSrc, pos)
			warnings[#warnings + 1] = {
				severity = 'warning', code = 'empty-sigil',
				message = string.format("empty '%s%s' does nothing", opener, closer),
				labels = { { line = line, col = col, len = #opener + #t.value + #closer, primary = true } },
				help = braced and 'remove it, or put a variable name inside'
					or 'remove it, or fill in an expression/command',
			}
		end
	end
	return warnings
end

-- collectDiagnostics runs the lexer + classifier diagnostics, plus the
-- empty-sigil warning, over one fragment (the top-level source when base==0,
-- otherwise a nested anonymous-function body), resolving everything to
-- absolute positions against topSrc as it goes.
local function collectDiagnostics(fragSrc, base, cdiags)
	local toks = lexer.tokenize(fragSrc)
	for _, d in ipairs(lexer.diagnose(toks, fragSrc)) do
		diagnostics[#diagnostics + 1] = shiftDiag(d, base)
	end
	for _, d in ipairs(cdiags or {}) do
		diagnostics[#diagnostics + 1] = shiftDiag(d, base)
	end
	for _, w in ipairs(emptySigilWarnings(toks, base)) do
		diagnostics[#diagnostics + 1] = w
	end
end

--- Returns and clears the diagnostics collected by the most recent
--- M.rewrite/M.chunk call (errors and warnings together).
--- @return table list of diagnostics (see diagnostic.lua for the shape)
function M.takeDiagnostics()
	local d = diagnostics
	diagnostics = {}
	return d
end

local function makeLineFinder(src)
	local breaks = {}
	for p in src:gmatch('()\n') do breaks[#breaks + 1] = p end
	return function(pos)
		local lo, hi = 0, #breaks
		while lo < hi do
			local mid = (lo + hi + 1) // 2
			if breaks[mid] < pos then lo = mid else hi = mid - 1 end
		end
		return lo + 1
	end
end

-- padTo returns the blank lines needed so the next emitted text lands on the
-- same line `pos` is on in the original source. Never moves backwards.
-- When lineFinder is disabled (nil), `pos` belongs to a re-tokenized text
-- fragment (e.g. an anonymous function body in rewriteLua) whose positions
-- are relative to that fragment, not the top-level src lineFinder was built
-- from, so alignment is meaningless there; fall back to the old
-- always-newline join instead of risking a bogus same-line decision.
-- when lineFinder is disabled (nested fragment), there is no source line to
-- target, so emittedLine can't be resynced via lineFinder(pos) - instead it's
-- advanced by exactly the literal newlines being emitted (here, and via the
-- countNewlines bookkeeping in rewriteStatements/the block branch below), so
-- it stays a true count of the real generated line, just not aligned to source.
local function padTo(pos)
	if not pos then return '' end
	if not lineFinder then
		emittedLine = emittedLine + 1
		return '\n'
	end
	local target = lineFinder(pos)
	local pad = target > emittedLine and string.rep('\n', target - emittedLine) or ''
	emittedLine = math.max(emittedLine, target)
	return pad
end

local function advanceTo(pos)
	if lineFinder and pos then emittedLine = math.max(emittedLine, lineFinder(pos)) end
end

local function luaReplacement(t, base)
	local kind = t.kind
	if kind == k.ENV then return 'paper.env(' .. fmt(t.value) .. ')' end
	if kind == k.CAPTURE then return 'paper.capture(' .. fmt(t.value) .. ')' end
	if kind == k.RESULT then return 'paper.result(' .. fmt(t.value) .. ')' end
	if kind == k.RUN then return 'paper.exec(' .. fmt(t.value) .. ')' end
	-- the eval body starts 2 chars past the sigil opener (`@(`); shift base so
	-- positions found while re-tokenizing t.value resolve absolutely.
	if kind == k.EVAL then return '(' .. rewriteLua(t.value, base + t.pos + 1) .. ')' end
end

local function functionBody(toks, i)
	local j = i + 1
	while toks[j] and toks[j].kind ~= k.LPAREN and toks[j].kind ~= k.EOF do j = j + 1 end
	if not toks[j] or toks[j].kind ~= k.LPAREN then return nil end
	local depth = 0
	repeat
		if toks[j].kind == k.LPAREN then depth = depth + 1
		elseif toks[j].kind == k.RPAREN then depth = depth - 1 end
		j = j + 1
	until depth == 0 or not toks[j] or toks[j].kind == k.EOF
	local bodyStart = j
	local block, pendingDo = 1, false -- the function itself is one open block
	while toks[j] and toks[j].kind ~= k.EOF do
		if toks[j].kind == k.KEYWORD then
			local v = toks[j].value
			if v == 'function' or v == 'if' or v == 'repeat' then
				block = block + 1
			elseif v == 'for' or v == 'while' then
				block = block + 1
				pendingDo = true
			elseif v == 'do' then
				if pendingDo then pendingDo = false else block = block + 1 end
			elseif v == 'end' or v == 'until' then
				block = block - 1
				if block == 0 then return bodyStart, j end
			end
		end
		j = j + 1
	end
	return nil
end

function rewriteLua(text, base)
	base = base or 0
	local toks = lexer.tokenize(text)
	local out, cursor, i = {}, 1, 1
	while toks[i] and toks[i].kind ~= k.EOF do
		local t = toks[i]
		if operators[t.kind] then
			out[#out + 1] = text:sub(cursor, t.pos - 1)
			out[#out + 1] = luaReplacement(t, base)
			cursor = t.pos + opSourceLen(text, t)
			i = i + 1
		elseif t.kind == k.KEYWORD and t.value == 'function' then
			local bodyStart, endIdx = functionBody(toks, i)
			if bodyStart and endIdx then
				out[#out + 1] = text:sub(cursor, toks[bodyStart].pos - 1) -- `function(...)`
				local body = text:sub(toks[bodyStart].pos, toks[endIdx].pos - 1)
				-- `body` is re-tokenized from position 1, so its positions are
				-- relative to this fragment, not the top-level src lineFinder
				-- was built from; disable alignment for this nested render, but
				-- shift base so absolute source positions still resolve, and
				-- advance emittedLine by the real newlines this nested render
				-- emits (rather than freezing it) so generated-line tracking -
				-- and anchors recorded inside the nested call - stay accurate
				-- for everything that follows.
				local nestedBase = base + toks[bodyStart].pos - 1
				local savedFinder, savedLine = lineFinder, emittedLine
				lineFinder = nil
				local bodyStmts, _, bodyDiags = classifier.parse(body)
				collectDiagnostics(body, nestedBase, bodyDiags)
				local nested = rewriteStatements(bodyStmts, nestedBase)
				local piece = '\n' .. nested .. '\nend'
				out[#out + 1] = piece
				lineFinder = savedFinder
				emittedLine = savedLine + countNewlines(piece)
				cursor = toks[endIdx].pos + 3 -- past `end`
				i = endIdx + 1
			else
				i = i + 1
			end
		else
			i = i + 1
		end
	end
	out[#out + 1] = text:sub(cursor)
	return table.concat(out)
end

local function rewriteCommand(text, base)
	base = base or 0
	local parts, cursor, spliced = {}, 1, false
	local function lit(s) if s ~= '' then parts[#parts + 1] = fmt(s) end end
	for _, t in ipairs(lexer.tokenize(text)) do
		if t.kind == k.EVAL then
			spliced = true
			lit(text:sub(cursor, t.pos - 1))
			parts[#parts + 1] = 'paper.quote(' .. rewriteLua(t.value, base + t.pos + 1) .. ')'
			cursor = t.pos + opSourceLen(text, t)
		end
	end
	if not spliced then return fmt(text) end
	lit(text:sub(cursor))
	if #parts == 0 then return fmt('') end
	return table.concat(parts, ' .. ')
end

local function funcState(text, base)
	base = base or 0
	local toks = lexer.tokenize(text)
	local i = 2
	while toks[i] and (toks[i].kind == k.DOT or toks[i].kind == k.COLON) do
		i = i + 2
	end

	local placeholder = false
	for _, t in ipairs(toks) do
		if t.kind == k.NAME and t.value == '_' then
			placeholder = true
			break
		end
	end

	if placeholder then
		local out, cursor = {}, 1
		for _, t in ipairs(toks) do
			if t.kind == k.NAME and t.value == '_' then
				out[#out + 1] = text:sub(cursor, t.pos - 1)
				out[#out + 1] = '__inp'
				cursor = t.pos + 1
			end
		end
		out[#out + 1] = text:sub(cursor)
		-- `_` -> `__inp` shifts everything after it, so positions found while
		-- re-tokenizing this reconstructed text no longer correspond 1:1 to
		-- `text`'s; base is kept as the closest still-useful approximation
		-- (this statement's span) rather than threading a second translation
		-- table through for a placeholder body's internal column.
		return 'function(__inp) return ' .. rewriteLua(table.concat(out), base) .. ' end'
	end

	local head = text:sub(1, (toks[i] and toks[i].pos - 1) or #text):gsub('%s+$', '')
	local rest = toks[i] and text:sub(toks[i].pos):gsub('%s+$', '') or ''

	local args
	if rest == '' then
		args = '__inp'
	elseif rest:sub(1, 1) == '(' then
		local inner = rest:sub(2, -2)
		args = inner == '' and '__inp' or ('__inp, ' .. inner)
	else
		args = '__inp, ' .. rest
	end

	return 'function(__inp) return ' .. rewriteLua(head, base) .. '(' .. args .. ') end'
end

local function pipeStageExpr(text, base)
	base = base or 0
	local trimmed = text:gsub('^%s+', '')
	base = base + (#text - #trimmed) -- account for the stripped leading whitespace
	text = trimmed:gsub('%s+$', '')
	local toks = lexer.tokenize(text)
	if toks[1] and toks[1].kind == k.NAME then
		local j = 2
		while toks[j] and (toks[j].kind == k.DOT or toks[j].kind == k.COLON) do j = j + 2 end
		if toks[j] and toks[j].kind ~= k.EOF then
			return funcState(text, base) -- has trailing args -> closure
		end
	end
	return rewriteLua(text, base)
end

local function thruChain(text, base)
	base = base or 0
	local cuts, depth, block, pendingDo = {}, 0, 0, false
	for _, t in ipairs(lexer.tokenize(text)) do
		if opens[t.kind] then depth = depth + 1
		elseif closes[t.kind] then depth = depth - 1
		elseif t.kind == k.PIPE and depth == 0 and block == 0 then cuts[#cuts + 1] = t.pos
		elseif t.kind == k.KEYWORD then
			local v = t.value
			if v == 'function' or v == 'if' or v == 'repeat' then block = block + 1
			elseif v == 'for' or v == 'while' then block = block + 1; pendingDo = true
			elseif v == 'do' then if pendingDo then pendingDo = false else block = block + 1 end
			elseif v == 'end' or v == 'until' then block = block - 1 end
		end
	end
	local segs, prev = {}, 1
	for _, cut in ipairs(cuts) do segs[#segs + 1] = { text:sub(prev, cut - 1), prev }; prev = cut + 1 end
	segs[#segs + 1] = { text:sub(prev), prev }
	local parts = {}
	for i, seg in ipairs(segs) do
		local s, relStart = seg[1], seg[2]
		parts[#parts + 1] = (i == 1) and rewriteLua(s, base) or pipeStageExpr(s, base + relStart - 1)
	end
	return 'paper.thru(' .. table.concat(parts, ', ') .. ')'
end

local function topLevelAssignRel(stmt)
	local depth, block, pendingDo = 0, 0, false
	for _, t in ipairs(stmt.tokens) do
		if opens[t.kind] then depth = depth + 1
		elseif closes[t.kind] then depth = depth - 1
		elseif t.kind == k.ASSIGN and depth == 0 and block == 0 then
			return t.pos - stmt.startPos + 1
		elseif t.kind == k.KEYWORD then
			local v = t.value
			if v == 'function' or v == 'if' or v == 'repeat' then block = block + 1
			elseif v == 'for' or v == 'while' then block = block + 1; pendingDo = true
			elseif v == 'do' then if pendingDo then pendingDo = false else block = block + 1 end
			elseif v == 'end' or v == 'until' then block = block - 1 end
		end
	end
	return nil
end

local function emitStatement(stmt, base)
	base = base or 0
	local stmtBase = base + stmt.startPos - 1
	local ty = stmt.type
	if ty == 'lua' then
		-- an all-Lua pipe: `local p = a | b` => `local p = paper.thru(a, <b>)`,
		-- a bare `a | b` => display the result.
		if stmt.thru then
			local arel = topLevelAssignRel(stmt)
			if arel then
				return stmt.src:sub(1, arel) .. ' ' .. thruChain(stmt.src:sub(arel + 1), stmtBase + arel)
			end
			return 'paper.show(' .. thruChain(stmt.src, stmtBase) .. ')'
		end
		-- a standalone `@(expr)` statement displays its value
		local t = stmt.tokens
		if #t == 1 and t[1].kind == k.EVAL then
			return 'paper.show(' .. rewriteLua(t[1].value, base + t[1].pos + 1) .. ')'
		end
		return rewriteLua(stmt.src, stmtBase)
	elseif ty == 'command' then
		return 'paper.exec(' .. rewriteCommand(stmt.src, stmtBase) .. ')'
	elseif ty == 'dispatch' then
		-- the string arg is everything after the head name in the span
		local rel = stmt.tokens[2].pos - stmt.startPos + 1
		local argSrc = (stmt.src:sub(rel):gsub('%s+$', ''))
		return 'paper.dispatch(' .. fmt(stmt.head) .. ', { ' .. argSrc .. ' }, '
			.. rewriteCommand(stmt.src, stmtBase) .. ')'
	elseif ty == 'pipeline' then
		local stages = {}
		for _, st in ipairs(stmt.stages) do
			local sBase = base + (st.pos or stmt.startPos)
			if st.kind == 'shell' then
				stages[#stages + 1] = '{ shell = ' .. rewriteCommand(st.src, sBase) .. ' }'
			elseif st.kind == 'func' then
				stages[#stages + 1] = '{ lua = ' .. funcState(st.src, sBase) .. ' }'
			else
				stages[#stages + 1] = '{ lua = ' .. rewriteLua(st.expr, sBase) .. ' }'
			end
		end
		return 'paper.pipeline({ ' .. table.concat(stages, ', ') .. ' })'
	elseif ty == 'func' then
		return funcState(stmt.src, stmtBase)
	elseif ty == 'block' then
		local out = {}
		for idx, p in ipairs(stmt.parts) do
			if p.span ~= nil then
				local pad = padTo(p.pos)
				local sep = (pad == '' and idx > 1) and ';' or ''
				recordAnchor(base + (p.pos or 0), #p.span)
				local piece = rewriteLua(p.span, base)
				out[#out + 1] = pad .. sep .. piece
				if lineFinder then
					advanceTo(p.pos and (p.pos + #p.span - 1))
				else
					emittedLine = emittedLine + countNewlines(piece)
				end
			else
				-- a same-line empty/short body (`if x then y end`) needs a `;`
				-- to separate it from the header span it's glued to; a body on
				-- its own line is already separated by padTo's newline.
				local sep = ''
				if p.body[1] and lineFinder and lineFinder(p.body[1].startPos) == emittedLine then
					sep = ';'
				end
				out[#out + 1] = sep .. rewriteStatements(p.body, base)
			end
		end
		return table.concat(out)
	end
end

function rewriteStatements(stmts, base)
	base = base or 0
	local out = {}
	for idx, s in ipairs(stmts) do
		local pad = padTo(s.startPos)
		local sep = (pad == '' and idx > 1) and ';' or ''
		recordAnchor(base + s.startPos, s.endPos - s.startPos + 1)
		local piece = emitStatement(s, base)
		out[#out + 1] = pad .. sep .. piece
		if lineFinder then
			advanceTo(s.endPos)
		else
			emittedLine = emittedLine + countNewlines(piece)
		end
	end
	return table.concat(out)
end

function M.rewrite(src)
	lineFinder = makeLineFinder(src)
	emittedLine = 1
	anchors = {}
	diagnostics = {}
	topSrc = src
	local stmts, _, diags = classifier.parse(src)
	collectDiagnostics(src, 0, diags)
	return rewriteStatements(stmts, 0)
end

local CHUNK_PREFIX = "local paper = require 'nature.paperbush.runtime'; "

function M.chunk(src)
	return CHUNK_PREFIX .. M.rewrite(src)
end

--- The byte length of the prefix M.chunk adds before the rewritten code.\
--- @return number
function M.chunkPrefixLen()
	return #CHUNK_PREFIX
end

return M
