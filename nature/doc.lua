-- @module doc
-- command-line doc rendering
-- The doc module contains a small set of functions
-- used by the Greenhouse pager to render parts of the documentation pages.
-- This is only documented for the sake of it. It's only intended use
-- is by the Greenhouse pager.
local lunacolors = require 'lunacolors'

local doc = {}

--- Performs basic Lua code highlighting.
--- @param text string Code/text to do highlighting on.
function doc.highlight(text)
	return text:gsub('\'.-\'', lunacolors.yellow)
	--:gsub('%-%- .-', lunacolors.black)
end

--- Assembles and renders a code block. This returns
--- the supplied text based on the number of command line columns,
--- and styles it to resemble a code block.
--- @param text string
function doc.renderCodeBlock(text)
	local longest = 0
	local lines = string.split(text:gsub('\t', '    '), '\n')

	for i, line in ipairs(lines) do
		local len = line:len()
		if len > longest then longest = len end
	end

	for i, line in ipairs(lines) do
		lines[i] = lunacolors.format('{greyBg}' .. ' ' .. doc.highlight(line:sub(0, longest))
		.. string.rep(' ', longest - line:len()) .. ' ')
	end

	return '\n' .. lunacolors.format('{greyBg}' .. table.concat(lines, '\n')) .. '\n'
end

--- Renders an info block. An info block is a block of text with
--- an icon and styled text block.
--- @param type string Type of info block. The only one specially styled is the `warning`.
--- @param text string
function doc.renderInfoBlock(type, text)
	local longest = 0
	local lines = string.split(text:gsub('\t', '    '), '\n')

	for i, line in ipairs(lines) do
		local len = line:len()
		if len > longest then longest = len end
	end

	for i, line in ipairs(lines) do
		lines[i] = ' ' .. doc.highlight(line:sub(0, longest))
		.. string.rep(' ', longest - line:len()) .. ' '
	end

	local heading
	if type == 'warning' then
		heading = lunacolors.yellowBg(lunacolors.black(' ⚠ Warning '))
	end
	return '\n' .. heading .. '\n' .. lunacolors.format('{greyBg}' .. table.concat(lines, '\n')) .. '\n'
end
return doc
