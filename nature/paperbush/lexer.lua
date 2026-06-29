local lexer = {}

local kinds = {
	EOF = 'EOF',
	ILLEGAL = 'ILLEGAL',

	NAME = 'NAME',
	KEYWORD = 'KEYWORD',
	NUMBER = 'NUMBER',
	STRING = 'STRING',
	LONGSTRING = 'LONGSTRING',
	COMMENT = 'COMMENT',

	-- sigils (Lua<->shell)
	CAPTURE = 'CAPTURE', -- $(
	RESULT = 'RESULT',   -- !(
	RUN = 'RUN',         -- $[
	EVAL = 'EVAL',   -- @(
	ENV = 'ENV',         -- $NAME / ${NAME}

	ASSIGN = 'ASSIGN',     -- =
	COMMA = 'COMMA',       -- ,
	DOT = 'DOT',           -- .
	COLON = 'COLON',       -- :
	LPAREN = 'LPAREN',     -- (
	RPAREN = 'RPAREN',     -- )
	LBRACE = 'LBRACE',     -- {
	RBRACE = 'RBRACE',     -- }
	LBRACKET = 'LBRACKET', -- [
	RBRACKET = 'RBRACKET', -- ]

	PIPE = 'PIPE',       -- |
	SEMI = 'SEMI',       -- ;
	NEWLINE = 'NEWLINE',

	-- pass-through (ran by either interpreter, lua or shell)
	OP = 'OP',
}
lexer.kinds = kinds

-- Lua keywords
local keywords = {
	['and'] = true, ['break'] = true, ['do'] = true, ['else'] = true,
	['elseif'] = true, ['end'] = true, ['false'] = true, ['for'] = true,
	['function'] = true, ['goto'] = true, ['if'] = true, ['in'] = true,
	['local'] = true, ['nil'] = true, ['not'] = true, ['or'] = true,
	['repeat'] = true, ['return'] = true, ['then'] = true, ['true'] = true,
	['until'] = true, ['while'] = true,
}
lexer.keywords = keywords

local single = {
	['='] = kinds.ASSIGN, [','] = kinds.COMMA, ['.'] = kinds.DOT,
	[':'] = kinds.COLON, ['('] = kinds.LPAREN, [')'] = kinds.RPAREN,
	['{'] = kinds.LBRACE, ['}'] = kinds.RBRACE, [']'] = kinds.RBRACKET,
	['|'] = kinds.PIPE, [';'] = kinds.SEMI,
}

-- sigil openers -> { kind, openChar, closeChar }
local sigils = {
	['$('] = { kinds.CAPTURE, '(', ')' },
	['!('] = { kinds.RESULT, '(', ')' },
	['@('] = { kinds.EVAL, '(', ')' },
	['$['] = { kinds.RUN, '[', ']' },
}

local Lexer = {}
Lexer.__index = Lexer

function lexer.new(src)
	return setmetatable({
		src = src,
		pos = 1,
		line = 1,
		col = 0,
	}, Lexer)
end

function Lexer:read()
	local c = self.src:sub(self.pos, self.pos)
	self.pos = self.pos + 1
	if c == '\n' then
		self.line = self.line + 1
		self.col = 0
	else
		self.col = self.col + 1
	end
	return c
end

function Lexer:peek() return self.src:sub(self.pos, self.pos) end
function Lexer:peek2() return self.src:sub(self.pos, self.pos + 1) end

function Lexer:mark() return { pos = self.pos, line = self.line, col = self.col } end
function Lexer:reset(m) self.pos, self.line, self.col = m.pos, m.line, m.col end

function Lexer:skipSpace()
	while true do
		local c = self:peek()
		if c == ' ' or c == '\t' or c == '\r' then
			self:read()
		else
			return
		end
	end
end

-- next() returns the next token: { kind, value, pos, line, col }
function Lexer:next()
	self:skipSpace()

	local m = self:mark()
	local function tok(kind, value)
		return { kind = kind, value = value, pos = m.pos, line = m.line, col = m.col + 1 }
	end

	local c = self:peek()

	if c == '' then
        self:read()
        return tok(kinds.EOF, '')
    end
	if c == '\n' then
        self:read()
        return tok(kinds.NEWLINE, '\n')
    end

	-- only treat `--` as a comment when
	-- followed by whitespace/EOL/`[`, so shell flags work
	if self:peek2() == '--' then
		local after = self.src:sub(self.pos + 2, self.pos + 2)
		if after == '[' then
			self:read()
			self:read()
			local level = self:tryLongBracket()
			if level then
				local body, incomplete = self:scanLongBracket(level)
				local t = tok(kinds.COMMENT, body)
				t.opener = '--[' .. string.rep('=', level) .. '['
				t.closer = ']' .. string.rep('=', level) .. ']'
				if incomplete then t.incomplete = true end
				return t
			end
			return tok(kinds.COMMENT, self:scanComment()) -- `--[x` is a line comment
		elseif after == '' or after == ' ' or after == '\t' or after == '\r' or after == '\n' then
			self:read(); self:read()
			return tok(kinds.COMMENT, self:scanComment())
		end
	end

	-- long strings: [[ ]] / [=[ ]=]
	if c == '[' then
		local level = self:tryLongBracket()
		if level then
			local body, incomplete = self:scanLongBracket(level)
			local t = tok(kinds.LONGSTRING, body)
			t.opener = '[' .. string.rep('=', level) .. '['
			t.closer = ']' .. string.rep('=', level) .. ']'
			if incomplete then t.incomplete = true end
			return t
		end
		self:read()
		return tok(kinds.LBRACKET, '[')
	end

	-- short strings
	if c == '"' or c == "'" then
		self:read()
		local lit, err = self:scanString(c)
		local t = tok(kinds.STRING, lit)
		t.quote = c
		if err then t.incomplete = true end
		return t
	end

	local sg = sigils[self:peek2()]
	if sg then
		local openTxt = self:peek2()
		self:read(); self:read()
		local body, incomplete = self:scanBalanced(sg[2], sg[3])
		local t = tok(sg[1], body)
		t.opener, t.closer = openTxt, sg[3]
		if incomplete then t.incomplete = true end
		return t
	end

	-- env refs: $NAME / ${NAME}. A `$` followed by neither is an operator
	if c == '$' then
		local nx = self.src:sub(self.pos + 1, self.pos + 1)
		if nx == '{' or nx:match('[%a_]') then
			self:read()
			local braced = nx == '{'
			local name, incomplete = self:scanEnv()
			local t = tok(kinds.ENV, name)
			if braced then t.opener, t.closer = '${', '}' end
			if incomplete then t.incomplete = true end
			return t
		end
	end

	-- names / keywords
	if c:match('[%a_]') then
		local ident = self:scanIdent()
		return tok(keywords[ident] and kinds.KEYWORD or kinds.NAME, ident)
	end

	-- numbers, including the leading-dot form (`.5`).
	if c:match('%d') or (c == '.' and self.src:sub(self.pos + 1, self.pos + 1):match('%d')) then
		return tok(kinds.NUMBER, self:scanNumber())
	end

	-- or operator
	if c == '|' and self:peek2() == '||' then
		self:read(); self:read()
		return tok(kinds.OP, '||')
	end

	if single[c] then
		self:read()
		return tok(single[c], c)
	end

	local op = self:scanOp()
	if op ~= '' then
		return tok(kinds.OP, op)
	end

	-- unrecognized byte
	return tok(kinds.ILLEGAL, self:read())
end

function Lexer:scanIdent()
	local start = self.pos
	while self:peek():match('[%w_]') do self:read() end
	return self.src:sub(start, self.pos - 1)
end

function Lexer:scanNumber()
	local start = self.pos
	if self:peek() == '0' then
		self:read()
		local x = self:peek()
		if x == 'x' or x == 'X' then
			self:read()
			while self:peek():match('[%x%.]') do self:read() end
			if self:peek():match('[pP]') then
				self:read()
				if self:peek():match('[%+%-]') then self:read() end
				while self:peek():match('%d') do self:read() end
			end
			return self.src:sub(start, self.pos - 1)
		end
	end
	while self:peek():match('[%d%.]') do self:read() end
	if self:peek():match('[eE]') then
		self:read()
		if self:peek():match('[%+%-]') then self:read() end
		while self:peek():match('%d') do self:read() end
	end
	return self.src:sub(start, self.pos - 1)
end

function Lexer:scanString(quote)
	local buf = {}
	while true do
		local c = self:read()
		if c == '' or c == '\n' then
			return table.concat(buf), 'unterminated string'
		elseif c == '\\' then
			local n = self:read()
			if n == '' then return table.concat(buf), 'unterminated string' end
			buf[#buf + 1] = c
			buf[#buf + 1] = n
		elseif c == quote then
			return table.concat(buf)
		else
			buf[#buf + 1] = c
		end
	end
end

-- scanBalanced reads a balanced `open`..`close` group body (the opening tok already consumed)
function Lexer:scanBalanced(open, close)
	local buf = {}
	local depth = 1
	while true do
		local c = self:read()
		if c == '' then
			return table.concat(buf), true
		elseif c == '\\' then
			local n = self:read()
			buf[#buf + 1] = c
			if n == '' then return table.concat(buf), true end
			buf[#buf + 1] = n
		elseif c == '"' or c == "'" then
			local s, err = self:scanString(c)
			buf[#buf + 1] = c
			buf[#buf + 1] = s
			if err then return table.concat(buf), true end
			buf[#buf + 1] = c
		elseif c == open then
			depth = depth + 1
			buf[#buf + 1] = c
		elseif c == close then
			depth = depth - 1
			if depth == 0 then return table.concat(buf), false end
			buf[#buf + 1] = c
		else
			buf[#buf + 1] = c
		end
	end
end

function Lexer:tryLongBracket()
	local m = self:mark()
	if self:read() ~= '[' then self:reset(m); return nil end
	local level = 0
	while self:peek() == '=' do self:read(); level = level + 1 end
	if self:read() == '[' then return level end
	self:reset(m)
	return nil
end

function Lexer:scanLongBracket(level)
	local buf = {}
	while true do
		local c = self:read()
		if c == '' then return table.concat(buf), true end -- unterminated
		if c == ']' then
			local m = self:mark()
			local eqs = 0
			while self:peek() == '=' do self:read(); eqs = eqs + 1 end
			if eqs == level and self:peek() == ']' then
				self:read()
				return table.concat(buf), false
			end
			self:reset(m)
			buf[#buf + 1] = c
		else
			buf[#buf + 1] = c
		end
	end
end

function Lexer:scanComment()
	local start = self.pos
	while true do
		local c = self:peek()
		if c == '' or c == '\n' then break end
		self:read()
	end
	return self.src:sub(start, self.pos - 1)
end

function Lexer:scanEnv()
	if self:peek() == '{' then
		self:read()
		local start = self.pos
		while true do
			local c = self:peek()
			if c == '' or c == '}' then break end
			self:read()
		end
		local name = self.src:sub(start, self.pos - 1)
		if self:peek() == '}' then
			self:read()
			return name, false
		end
		return name, true -- unterminated ${
	end
	local start = self.pos
	while self:peek():match('[%w_]') do self:read() end
	return self.src:sub(start, self.pos - 1), false
end

function Lexer:scanOp()
	local start = self.pos
	while self:peek():match('[%+%-%*/%%%^#~<>=&!]') do self:read() end
	return self.src:sub(start, self.pos - 1)
end

function lexer.tokenize(src)
	local lx = lexer.new(src)
	local toks = {}
	while true do
		local t = lx:next()
		toks[#toks + 1] = t
		if t.kind == kinds.EOF then break end
	end
	return toks
end

function lexer.dump(src)
	local toks = type(src) == 'table' and src or lexer.tokenize(src)
	local lines = {}
	for _, t in ipairs(toks) do
		lines[#lines + 1] = string.format(
			'%3d:%-3d  %-10s %s%s',
			t.line or 0, t.col or 0, t.kind,
			('%q'):format(t.value),
			t.incomplete and '  <incomplete>' or ''
		)
	end
	local out = table.concat(lines, '\n')
	return out
end

-- isValidNumber re-checks a scanned NUMBER token's text against Lua's actual
-- number grammar. scanNumber is deliberately permissive (it just gobbles
-- digits/dots/hex/exponent runs), so it happily produces tokens like `1..2` or
-- `0x` that Lua's own lexer would reject; this catches those before they ever
-- reach load().
local function isValidNumber(s)
	if s:match('^0[xX]') then
		local body = s:sub(3)
		local mantissa, exp = body:match('^([^pP]*)(.*)$')
		if exp ~= '' and not exp:match('^[pP][%+%-]?%d+$') then return false end
		if mantissa == '' then return false end
		if not mantissa:match('^%x*%.?%x*$') then return false end
		return true
	end
	local mantissa, exp = s:match('^([^eE]*)(.*)$')
	if exp ~= '' and not exp:match('^[eE][%+%-]?%d+$') then return false end
	if mantissa == '' or mantissa == '.' then return false end
	if not mantissa:match('^%d*%.?%d*$') then return false end
	return true
end
lexer.isValidNumber = isValidNumber

local sigilNoun = {
	[kinds.CAPTURE] = 'capture', [kinds.RESULT] = 'result',
	[kinds.EVAL] = 'eval', [kinds.RUN] = 'run',
}

--- diagnose scans an already-tokenized list for lexical errors: illegal bytes,
--- unterminated strings/long-brackets/comments/sigils, and malformed numbers.
--- Every position comes straight from the tokens (no Lua round-trip), so these
--- are exact. `src` is needed only to compute the "ran out of input" position
--- (one past the last character) for unterminated constructs.
--- @param toks table token list from lexer.tokenize
--- @param src string the original source the tokens were lexed from
--- @return table list of diagnostics (see diagnostic.lua for the shape)
function lexer.diagnose(toks, src)
	local diags = {}

	local eofLine, lastNL = 1, 0
	for p in src:gmatch('()\n') do lastNL = p; eofLine = eofLine + 1 end
	local eofCol = #src - lastNL + 1
	local eofPos = #src + 1

	for _, t in ipairs(toks) do
		if t.kind == kinds.ILLEGAL then
			diags[#diags + 1] = {
				severity = 'error', code = 'illegal-char',
				message = string.format('unexpected character %q', t.value),
				labels = { { line = t.line, col = t.col, pos = t.pos, len = 1, text = 'not valid here', primary = true } },
				help = 'remove this character, or quote it if it is meant to be literal text',
			}
		elseif t.incomplete and t.kind == kinds.STRING then
			diags[#diags + 1] = {
				severity = 'error', code = 'unterminated',
				message = 'unterminated string',
				labels = {
					{ line = t.line, col = t.col, pos = t.pos, len = 1, text = 'string starts here', primary = true },
					{ line = eofLine, col = eofCol, pos = eofPos, len = 1,
						text = string.format("expected closing %s before end of input", ('%q'):format(t.quote or '"')) },
				},
				help = 'add the matching quote, or escape it with \\ if it is meant to be literal',
			}
		elseif t.incomplete and (t.kind == kinds.LONGSTRING or t.kind == kinds.COMMENT) then
			diags[#diags + 1] = {
				severity = 'error', code = 'unterminated',
				message = string.format('unterminated %s', t.kind == kinds.COMMENT and 'long comment' or 'long string'),
				labels = {
					{ line = t.line, col = t.col, pos = t.pos, len = #t.opener, text = 'opened here', primary = true },
					{ line = eofLine, col = eofCol, pos = eofPos, len = 1,
						text = string.format("expected '%s' before end of input", t.closer) },
				},
				help = string.format("close it with '%s'", t.closer),
			}
		elseif t.incomplete and sigilNoun[t.kind] then
			diags[#diags + 1] = {
				severity = 'error', code = 'unterminated',
				message = string.format("unclosed '%s' %s", t.opener, sigilNoun[t.kind]),
				labels = {
					{ line = t.line, col = t.col, pos = t.pos, len = #t.opener, text = 'opened here', primary = true },
					{ line = eofLine, col = eofCol, pos = eofPos, len = 1,
						text = string.format("expected '%s' before end of input", t.closer) },
				},
				help = string.format("close it with '%s'", t.closer),
			}
		elseif t.incomplete and t.kind == kinds.ENV then
			diags[#diags + 1] = {
				severity = 'error', code = 'unterminated',
				message = "unclosed '${' environment reference",
				labels = {
					{ line = t.line, col = t.col, pos = t.pos, len = #(t.opener or '${'), text = 'opened here', primary = true },
					{ line = eofLine, col = eofCol, pos = eofPos, len = 1, text = "expected '}' before end of input" },
				},
				help = "close it with '}'",
			}
		elseif t.kind == kinds.NUMBER and not isValidNumber(t.value) then
			diags[#diags + 1] = {
				severity = 'error', code = 'malformed-number',
				message = string.format("malformed number near '%s'", t.value),
				labels = { { line = t.line, col = t.col, pos = t.pos, len = #t.value, text = 'not a valid number literal', primary = true } },
				help = 'a number may have at most one decimal point and one exponent (e/E, or p/P for hex floats)',
			}
		end
	end

	return diags
end

return lexer
