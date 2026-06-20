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
	SPLICE = 'SPLICE',   -- @(
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
	['@('] = { kinds.SPLICE, '(', ')' },
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

	-- comments `--` line and `--[[ ]]` block
	if self:peek2() == '--' then
		self:read()
        self:read()
		if self:peek() == '[' then
			local level = self:tryLongBracket()
			if level then
				local body, incomplete = self:scanLongBracket(level)
				local t = tok(kinds.COMMENT, body)
				if incomplete then t.incomplete = true end
				return t
			end
		end
		return tok(kinds.COMMENT, self:scanComment())
	end

	-- long strings: [[ ]] / [=[ ]=]
	if c == '[' then
		local level = self:tryLongBracket()
		if level then
			local body, incomplete = self:scanLongBracket(level)
			local t = tok(kinds.LONGSTRING, body)
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
		if err then t.incomplete = true end
		return t
	end

	local sg = sigils[self:peek2()]
	if sg then
		self:read(); self:read()
		local body, incomplete = self:scanBalanced(sg[2], sg[3])
		local t = tok(sg[1], body)
		if incomplete then t.incomplete = true end
		return t
	end

	-- env refs: $NAME / ${NAME}. A `$` followed by neither is an operator
	if c == '$' then
		local nx = self.src:sub(self.pos + 1, self.pos + 1)
		if nx == '{' or nx:match('[%a_]') then
			self:read()
			local name, incomplete = self:scanEnv()
			local t = tok(kinds.ENV, name)
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

return lexer
