local codegen = require 'nature.paperbush.codegen'

local M = {}

local function lineColAt(src, pos)
	local line, lastNL = 1, 0
	for p in src:gmatch('()\n') do
		if p >= pos then break end
		lastNL = p
		line = line + 1
	end
	return line, pos - lastNL
end

-- findAnchor returns the anchor with the largest genLine <= genLine: the
-- statement/span whose generated text is "open" at that line.
local function findAnchor(map, genLine)
	local best
	for _, a in ipairs(map or {}) do
		if a.genLine <= genLine and (not best or a.genLine > best.genLine) then best = a end
	end
	return best
end

local function isWordToken(token) return token:match('^[%w_]+$') ~= nil end

-- findAllOccurrences finds every place `token` appears in `span`, using a
-- word-boundary match for identifier/keyword tokens (so `x` doesn't match
-- inside `max`) and a plain literal search for punctuation/operator tokens
-- (word-boundary frontiers don't apply to them).
local function findAllOccurrences(span, token)
	local positions = {}
	if isWordToken(token) then
		local pat = '%f[%w_]' .. token .. '%f[^%w_]'
		local from = 1
		while true do
			local s, e = span:find(pat, from)
			if not s then break end
			positions[#positions + 1] = s
			from = e + 1
		end
	end
	if #positions == 0 then
		local escaped = token:gsub('%W', '%%%1')
		local from = 1
		while true do
			local s, e = span:find(escaped, from)
			if not s then break end
			positions[#positions + 1] = s
			from = e + 1
		end
	end
	return positions
end

-- findTokenInSpan locates `token` within `span`, picking the occurrence
-- closest to `wantOffset` (a same-frame offset derived from gopher-lua's own
-- column) when there's more than one and a hint is available.
local function findTokenInSpan(span, token, wantOffset)
	if not token or token == '' then return nil end
	local positions = findAllOccurrences(span, token)
	if #positions == 0 then return nil end
	if not wantOffset then return positions[1] end
	local best, bestDist = positions[1], math.abs(positions[1] - wantOffset)
	for i = 2, #positions do
		local dist = math.abs(positions[i] - wantOffset)
		if dist < bestDist then best, bestDist = positions[i], dist end
	end
	return best
end

-- extractLocator pulls the token Lua's own message already points at: either
-- a named-variable parenthetical (kept for forward-compat - PUC-Lua includes
-- this; this runtime currently doesn't) or a syntax-error "near" clause.
local function extractLocator(message)
	local _, name = message:match("%((%a+) '([^']*)'%)")
	if name then return name end
	return message:match("near '([^']*)'")
end

local translations = {
	{ pat = "^attempt to call a nil value$",
		build = function(_, _)
			return 'call-nil', "attempt to call a nil value",
				"check that this is defined and spelled correctly, and that you didn't mean to run it as a command"
		end },
	{ pat = "^attempt to index a nil value$",
		build = function(_, _)
			return 'index-nil', 'attempt to index a nil value',
				'check that this is assigned before this point'
		end },
	{ pat = "^attempt to perform arithmetic on a nil value$",
		build = function(_, _)
			return 'arith-nil', 'attempt to perform arithmetic on a nil value',
				'check that this holds a number here'
		end },
	{ pat = "^attempt to concatenate a nil value with a (%a+) value$",
		build = function(otherTy, _)
			return 'concat-nil', string.format('attempt to concatenate a nil value with a %s value', otherTy),
				'check that both sides are a string or number here'
		end },
	{ pat = "^attempt to call a nil value %((%a+) '([^']*)'%)$",
		build = function(kind, name)
			return 'call-nil', string.format("'%s' is not defined", name),
				kind == 'global'
					and string.format("if you meant to run a command, make sure '%s' is on PATH; if you meant a variable, check the spelling", name)
					or string.format("'%s' has no value assigned yet here", name)
		end },
	{ pat = "^attempt to index a nil value %(%a+ '([^']*)'%)$",
		build = function(name, _)
			return 'index-nil', string.format("'%s' is nil, so it has no fields", name),
				string.format("check that '%s' is assigned before this point", name)
		end },
	{ pat = "^attempt to compare (%a+) with (%a+)$",
		build = function(a, b)
			return 'compare-mismatch', string.format("can't compare %s with %s", a, b), nil
		end },
	{ pat = "^(.-) near '[^']*'$",
		build = function(desc, _)
			return 'syntax', desc, nil
		end },
}

local function translate(message)
	for _, t in ipairs(translations) do
		local c1, c2 = message:match(t.pat)
		if c1 then return t.build(c1, c2) end
	end
	return nil, message, nil
end

local function fromLua(err, source, map)
	if type(err) ~= 'string' then
		return { severity = 'error', code = 'lua-error', message = tostring(err), labels = {} }
	end

	local lineStr, colStr, message = err:match("^=?paperbush:(%d+):(%d+):%s*(.*)$")
	if not lineStr then
		lineStr, message = err:match('^=?paperbush:(%d+):%s*(.*)$')
	end
	if not lineStr then
		return { severity = 'error', code = 'lua-error', message = err, labels = {} }
	end

	local genLine = tonumber(lineStr)
	local code, friendly, help = translate(message)
	local anchor = findAnchor(map, genLine)
	if not anchor then
		return { severity = 'error', code = code or 'lua-error', message = friendly, help = help, labels = {} }
	end

	local span = source:sub(anchor.srcPos, anchor.srcPos + anchor.len - 1)
	local token = extractLocator(message)

	-- gopher-lua's column (load() errors only) is relative to the generated
	-- line, which only matches the source span's own offsets when nothing
	-- before it on that line has changed length (true for the common case:
	-- a statement alone on a fresh generated line). Used purely to
	-- disambiguate a token that appears more than once in the span (e.g. the
	-- two '=' in `local x = = 5`)
	local wantOffset
	if colStr then
		wantOffset = tonumber(colStr) - (genLine == 1 and codegen.chunkPrefixLen() or 0)
	end

	local rel = findTokenInSpan(span, token, wantOffset)
	local labelPos, labelLen
	if rel then
		labelPos, labelLen = anchor.srcPos + rel - 1, #token
	else
		labelPos, labelLen = anchor.srcPos, anchor.len
	end
	local line, col = lineColAt(source, labelPos)

	return {
		severity = 'error', code = code or 'lua-error', message = friendly, help = help,
		labels = { { line = line, col = col, len = labelLen, primary = true } },
	}
end

--- Maps a load() compile error back to a source diagnostic.
--- @param err any the second return of load() on failure
--- @param source string the original user-typed input
--- @param map table codegen.lastMap()'s anchor list for this same input
--- @return table a diagnostic (see diagnostic.lua)
function M.fromLoad(err, source, map)
	return fromLua(err, source, map)
end

--- Maps a pcall() runtime error back to a source diagnostic.
--- @param err any the second return of pcall() on failure
--- @param source string the original user-typed input
--- @param map table codegen.lastMap()'s anchor list for this same input
--- @return table a diagnostic (see diagnostic.lua)
function M.fromRuntime(err, source, map)
	return fromLua(err, source, map)
end

return M
