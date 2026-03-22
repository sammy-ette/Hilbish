local ansikit = require 'ansikit'
local bait = require 'bait'
local commander = require 'commander'
local hilbish = require 'hilbish'
local lunacolors = require 'lunacolors'
local terminal = require 'terminal'
local Greenhouse = require 'nature.greenhouse'
local Page = require 'nature.greenhouse.page'

commander.register('greenhouse', function(args, sinks)
	local gh = Greenhouse(sinks.out)

	local buffer = ''
	local display = ''
	local command = false
	local commands = {
		q = function()
			gh.keybinds['Ctrl-D'](gh)
		end,
		['goto'] = function(args)
			if not args[1] then
				return 'nuh uh'
			end
			gh:jump(tonumber(args[1]))
		end
	}

	function gh:resize()
		local size = terminal.size()
		self.region = {
			width = size.width,
			height = size.height - 2
		}
	end

	function gh:render()
		local workingPage = self.pages[self.curPage]
		local offset = self.offset
		if self.isSpecial then
			offset = self.specialOffset
			workingPage = self.specialPage
		end

		self.sink:write(ansikit.getCSI(self.region.height + 1 .. ';1', 'H'))
		if not self.isSpecial then
			self.sink:writeln(lunacolors.format(string.format('{grayBg} ↳ Page %d%s{reset}', self.curPage, workingPage.title and ' — ' .. workingPage.title .. ' ' or '')))
		end
		self.sink:write(buffer == '' and display or buffer)
	end
	function gh:input(c)
		-- command handling
		if c == ':' and not command then
			command = true
		end
		if c == 'Escape' then
			if command then
				command = false
				buffer = ''
			else
				if self.isSpecial then gh:special() end
			end
		elseif c == 'Backspace' then
			buffer = buffer:sub(0, -2)
			if buffer == '' then
				command = false
			else
				goto update
			end
		end

		if command then
			ansikit.showCursor()
			if buffer:match '^:' then buffer = buffer .. c else buffer = c end
		else
			ansikit.hideCursor()
		end

		::update::
		gh:update()
	end
	gh:resize()

	gh:keybind('Enter', function(self)
		if self.isSpecial then
			self:jump(self.specialPageIdx)
			self:special(false)
		else
			if buffer:len() < 2 then return end

			local splitBuf = string.split(buffer, " ")
			local command = commands[splitBuf[1]:sub(2)]
			if command then
				table.remove(splitBuf, 1)
				buffer = command(splitBuf) or ''
			end
			self:update()
		end
	end)

	if sinks['in'].pipe then
		local page = Page('stdin', sinks['in']:readAll())
		gh:addPage(page)
	end

	for _, name in ipairs(args) do
		local f <close> = io.open(name, 'r')
		if not f then
			sinks.err:writeln(string.format('could not open file %s', name))
		end
		local page = Page(name, f:read '*a')
		gh:addPage(page)
	end

	if #gh.pages == 0 then
		sinks.out:writeln [[greenhouse is the Hilbish pager library and command!
usage: greenhouse <file>...

example: greenhouse hello.md]]
		return 1
	end

	ansikit.hideCursor()
	gh:initUi()
end)
