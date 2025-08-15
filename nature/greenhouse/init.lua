-- @module greenhouse
-- Greenhouse is a simple text scrolling handler (pager) for terminal programs.
-- The idea is that it can be set a region to do its scrolling and paging
-- job and then the user can draw whatever outside it.
-- This reduces code duplication for the message viewer
-- and flowerbook.

local ansikit = require 'ansikit'
local lunacolors = require 'lunacolors'
local terminal = require 'terminal'
local Page = require 'nature.greenhouse.page'
local Object = require 'nature.object'

local Greenhouse = Object:extend()

function Greenhouse:new(sink)
	local size = terminal.size()
	self.region = size
	self.contents = nil -- or can be a table
	self.start = 1 -- where to start drawing from (should replace with self.region.y)
	self.offset = 1 -- vertical text offset
	self.horizOffset = 1
	self.sink = sink
	self.pages = {}
	self.curPage = 1
	self.step = {
		horizontal = 5,
		vertical = 1
	}
	self.separator = '─'
	self.keybinds = {
		['Up'] = function(self) self:scroll 'up' end,
		['Down'] = function(self) self:scroll 'down' end,
		['Left'] = function(self) self:scroll 'left' end,
		['Right'] = function(self) self:scroll 'right' end,
		['Ctrl-Left'] = self.previous,
		['Ctrl-Right'] = self.next,
		['Ctrl-N'] = function(self) self:toc(true) end,
		['Enter'] = function(self)
			if self.isSpecial then
				self:jump(self.specialPageIdx)
				self:special(false)
			end
		end,
		['Page-Down'] = function(self) self:scroll('down', {page = true}) end,
		['Page-Up'] = function(self) self:scroll('up', {page = true}) end
	}
	self.isSpecial = false
	self.specialPage = nil
	self.specialPageIdx = 1
	self.specialOffset = 1

	return self
end

function Greenhouse:addPage(page)
	table.insert(self.pages, page)
end

function Greenhouse:updateCurrentPage(text)
	local page = self.pages[self.curPage]
	page:setText(text)
end

local ansiPatters = {
	'\x1b%[%d+;%d+;%d+;%d+;%d+%w',
	'\x1b%[%d+;%d+;%d+;%d+%w',
	'\x1b%[%d+;%d+;%d+%w',
	'\x1b%[%d+;%d+%w',
	'\x1b%[%d+%w'
}

function Greenhouse:sub(str, offset, limit)
	local overhead = 0
	local function addOverhead(s)
		overhead = overhead + string.len(s)
	end

	local s = str
	for _, pat in ipairs(ansiPatters) do
		s = s:gsub(pat, addOverhead)
	end

	return s:sub(offset, utf8.offset(str, limit + overhead) or limit + overhead)
	--return s:sub(offset, limit + overhead)
end

function Greenhouse:draw()
	local workingPage = self.pages[self.curPage]
	local offset = self.offset
	if self.isSpecial then
		offset = self.specialOffset
		workingPage = self.specialPage
	end

	if workingPage.lazy and not workingPage.loaded then
		workingPage.initialize()
	end

	local lines = workingPage.lines
	self.sink:write(ansikit.getCSI(self.start .. ';1', 'H'))
	self.sink:write(ansikit.getCSI(2, 'J'))

	local writer = self.sink.writeln
	self.attributes = {}
	for i = offset, offset + self.region.height - 1 do
		local resetEnd = false
		if i > #lines then break end

		if i == offset + self.region.height - 1 then writer = self.sink.write end

		self.sink:write(ansikit.getCSI(self.start + i - offset .. ';1', 'H'))
		local line = lines[i]:gsub('{separator}', function() return self.separator:rep(self.region.width - 1) end)
		for _, pat in ipairs(ansiPatters) do
			line:gsub(pat, function(s)
				if s == lunacolors.formatColors.reset then
					self.attributes = {}
					resetEnd = true
				else
					--resetEnd = false
					--table.insert(self.attributes, s)
				end
			end)
		end

--[[
		if #self.attributes ~= 0 then
			for _, attr in ipairs(self.attributes) do
				--writer(self.sink, attr)
			end
		end
]]--

		self.sink:write(lunacolors.formatColors.reset)
		writer(self.sink, self:sub(line:gsub('\t', '        '), self.horizOffset, self.region.width + self.horizOffset))
		if resetEnd then
			self.sink:write(lunacolors.formatColors.reset)
		end
	end
	writer(self.sink, '\27[0m')
	self:render()
end

function Greenhouse:render()
end

function Greenhouse:scroll(direction, opts)
	opts = opts or {}

	if self.isSpecial then
		if direction == 'down' then
			self:next(true)
		elseif direction == 'up' then
			self:previous(true)
		end
		return
	end

	local lines = self.pages[self.curPage].lines

	local oldOffset = self.offset
	local oldHorizOffset = self.horizOffset
	local amount = self.step.vertical
	if opts.page then
		amount = self.region.height
	end

	if direction == 'down' then
		self.offset = math.min(self.offset + amount, math.max(1, #lines - self.region.height))
	elseif direction == 'up' then
		self.offset = math.max(self.offset - amount, 1)
	end

--[[
	if direction == 'left' then
		self.horizOffset = math.max(self.horizOffset - self.step.horizontal, 1)
	elseif direction == 'right' then
		self.horizOffset = self.horizOffset + self.step.horizontal
	end
]]--

	if self.offset ~= oldOffset then self:draw() end
	if self.horizOffset ~= oldHorizOffset then self:draw() end
end

function Greenhouse:update()
	self:resize()
	if self.isSpecial then
		self:updateSpecial()
	end

	self:draw()
end


function Greenhouse:special(val)
	self.isSpecial = val
	self:update()
end

function Greenhouse:toggleSpecial()
	self:special(not self.isSpecial)
end

--- This function will be called when the special page
--- is on and needs to be updated.
function Greenhouse:updateSpecial()
end

function Greenhouse:contents()
end

function Greenhouse:toc(toggle)
	if not self.isSpecial then
		self.specialPageIdx = self.curPage
	end
	if toggle then self.isSpecial = not self.isSpecial end
	-- Generate a special page for our table of contents
	local tocText = string.format([[
%s

]], lunacolors.cyan(lunacolors.bold '―― Table of Contents ――'))

	local genericPageCount = 1
	local contents = self:contents()
	if contents then
		for i, c in ipairs(contents) do
			local title = c.title
			if c.active then
				title = lunacolors.invert(title)
			end

			tocText = tocText .. title .. '\n'
		end
	else
		for i, page in ipairs(self.pages) do
			local title = page.title
			if title == 'Page' then
				title = 'Page #' .. genericPageCount
				genericPageCount = genericPageCount + 1
			end
			if i == self.specialPageIdx then
				title = lunacolors.invert(title)
			end

			tocText = tocText .. title .. '\n'
		end
	end
	self.specialPage = Page('TOC', tocText)
	function self:updateSpecial()
		self:toc()
	end
	self:draw()
end

function Greenhouse:resize()
	local size = terminal.size()
	self.region = size
end

function Greenhouse:next(special)
	local oldCurrent = special and self.specialPageIdx or self.curPage
	local pageIdx = math.min(oldCurrent + 1, #self.pages)

	if special then
		self.specialPageIdx = pageIdx
	else
		self.curPage = pageIdx
	end

	if pageIdx ~= oldCurrent then
		self.offset = 1
		self:update()
	end
end

function Greenhouse:previous(special)
	local oldCurrent = special and self.specialPageIdx or self.curPage
	local pageIdx = math.max(self.curPage - 1, 1)

	if special then
		self.specialPageIdx = pageIdx
	else
		self.curPage = pageIdx
	end

	if pageIdx ~= oldCurrent then
		self.offset = 1
		self:update()
	end
end

function Greenhouse:jump(idx)
	if idx ~= self.curPage then
		self.offset = 1
	end
	self.curPage = idx
	self:update()
end

function Greenhouse:keybind(key, callback)
	self.keybinds[key] = callback
end

function Greenhouse:input(char)
end

local function read()
	terminal.saveState()
	terminal.setRaw()
	local c = hilbish.editor.readChar()

	terminal.restoreState()
	return c
end

function Greenhouse:initUi()
	local ansikit = require 'ansikit'
	local bait = require 'bait'
	local commander = require 'commander'
	local hilbish = require 'hilbish'
	local terminal = require 'terminal'
	local Page = require 'nature.greenhouse.page'
	local done = false

	local function sigint()
		ansikit.clear()
		done = true
	end

	local function resize()
		self:update()
	end
	bait.catch('signal.sigint', sigint)

	bait.catch('signal.resize', resize)

	ansikit.screenAlt()
	ansikit.clear(true)
	self:draw()

	while not done do
		local c = read()
		self:keybind('Ctrl-Q', function()
			done = true
		end)
		self:keybind('Ctrl-D', function()
			done = true
		end)

		if self.keybinds[c] then
			self.keybinds[c](self)
		else
			self:input(c)
		end
	end

	ansikit.showCursor()
	ansikit.screenMain()

	self = nil
	bait.release('signal.sigint', sigint)
	bait.release('signal.resize', resize)

	ansikit.clear()
end

return Greenhouse
