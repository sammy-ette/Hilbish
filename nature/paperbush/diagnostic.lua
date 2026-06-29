-- A diagnostic is a plain table:
--   {
--     severity = 'error' | 'warning' | 'note',
--     code     = 'unbalanced',           -- optional short tag, shown as error[code]:
--     message  = "no '(' to close",
--     labels   = {                       -- one or more caret spans
--       { line = 1, col = 7, len = 1, text = "...", primary = true },
--     },
--     help     = "remove it, or ...",    -- optional, may contain newlines
--   }
-- col/len are 1-based byte columns into the source line

local lunacolors = require 'lunacolors'

local M = {}

local severityColor = { error = 'red', warning = 'yellow', note = 'cyan' }

local function painter(enabled)
	return function(name, s)
		if not enabled or s == nil then return s end
		local fn = lunacolors[name]
		return fn and fn(s) or s
	end
end

local TABSTOP = 4
local function expandTabs(s) return (s:gsub('\t', string.rep(' ', TABSTOP))) end

local function displayCol(line, col)
	if not col or col < 1 then return 1 end
	return #expandTabs(line:sub(1, col - 1)) + 1
end

local function splitLines(src)
	local lines = {}
	-- gmatch with [^\n]* skips trailing-empty handling; do it manually so a final
	-- line without a newline is still captured.
	local pos = 1
	while true do
		local nl = src:find('\n', pos, true)
		if not nl then
			lines[#lines + 1] = src:sub(pos)
			break
		end
		lines[#lines + 1] = src:sub(pos, nl - 1)
		pos = nl + 1
	end
	return lines
end

--- Render a diagnostic against its source, returning the framed string.
--- @param diag table
--- @param source string
--- @param opts table? { color = boolean } -- defaults to NO_COLOR detection
--- @return string
function M.render(diag, source, opts)
	opts = opts or {}
	local useColor = opts.color
	if useColor == nil then
		local nc = os.getenv('NO_COLOR')
		useColor = not (nc and nc ~= '')
	end
	local c = painter(useColor)

	local severity = diag.severity or 'error'
	local sevName = severityColor[severity] and severity or 'error'
	local sevColor = severityColor[sevName]
	local lines = splitLines(source)

	local labels = {}
	for _, l in ipairs(diag.labels or {}) do
		if l.line and l.line >= 1 and l.line <= #lines then
			labels[#labels + 1] = l
		end
	end
	local primary = labels[1]
	for _, l in ipairs(labels) do if l.primary then primary = l; break end end

	local maxLine = 1
	for _, l in ipairs(labels) do if l.line > maxLine then maxLine = l.line end end
	local gw = #tostring(maxLine)
	local function pad(n) return string.format('%' .. gw .. 'd', n) end
	local railBlank = c('grey', string.rep(' ', gw) .. ' │')

	local out = {}
	-- severity[code]: message
	local head = sevName
	if diag.code then head = head .. '[' .. diag.code .. ']' end
	out[#out + 1] = c(sevColor, c('bold', head)) .. c('bold', ': ' .. (diag.message or ''))

	local locLabel = primary or labels[1]
	if locLabel then
		local locStr = 'paperbush:' .. tostring(locLabel.line or '?')
		if locLabel.col then locStr = locStr .. ':' .. tostring(locLabel.col) end
		out[#out + 1] = c('grey', string.rep(' ', gw + 1) .. '┌─ ') .. locStr
	end

	if #labels == 0 then
		if diag.help then
			out[#out + 1] = railBlank
			out[#out + 1] = c('grey', string.rep(' ', gw) .. ' = ')
				.. c('cyan', c('bold', 'help: ')) .. diag.help
		end
		return table.concat(out, '\n')
	end

	out[#out + 1] = railBlank

	local byLine, order = {}, {}
	for _, l in ipairs(labels) do
		if not byLine[l.line] then byLine[l.line] = {}; order[#order + 1] = l.line end
		table.insert(byLine[l.line], l)
	end
	table.sort(order)

	for _, ln in ipairs(order) do
		local srcLine = lines[ln]
		out[#out + 1] = c('grey', pad(ln) .. ' │ ') .. expandTabs(srcLine)
		local group = byLine[ln]
		table.sort(group, function(a, b) return (a.col or 0) < (b.col or 0) end)
		for _, l in ipairs(group) do
			if l.col then
				local mark = l.primary and '^' or '~'
				local indent = string.rep(' ', displayCol(srcLine, l.col) - 1)
				local carets = string.rep(mark, math.max(l.len or 1, 1))
				local annot = l.text and (' ' .. l.text) or ''
				out[#out + 1] = c('grey', string.rep(' ', gw) .. ' │ ')
					.. indent .. c(sevColor, carets .. annot)
			end
		end
	end

	if diag.help then
		out[#out + 1] = railBlank
		local helpLines = splitLines(diag.help)
		for i, hl in ipairs(helpLines) do
			if i == 1 then
				out[#out + 1] = c('grey', string.rep(' ', gw) .. ' = ')
					.. c('cyan', c('bold', 'help: ')) .. hl
			else
				out[#out + 1] = string.rep(' ', gw) .. '       ' .. hl
			end
		end
	end

	return table.concat(out, '\n')
end

--- Render `diag` and write it to stderr (with a trailing newline).
--- @param diag table
--- @param source string
--- @param opts table?
function M.report(diag, source, opts)
	io.stderr:write(M.render(diag, source, opts) .. '\n')
end

return M
