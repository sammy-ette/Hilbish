local ansikit = require 'ansikit'
local commander = require 'commander'
local fs = require 'fs'
local lunacolors = require 'lunacolors'
local Greenhouse = require 'nature.greenhouse'
local Page = require 'nature.greenhouse.page'
local docfuncs = require 'nature.doc'

local function strip(text, ...)
	for _, pat in ipairs {...} do
		text = text:gsub(pat, '\n')
	end

	return text
end

local function transformHTMLandMD(text)
	return strip(text, '|||', '|%-%-%-%-|%-%-%-%-|')
	:gsub('|(.-)|(.-)|', function(entry1, entry2)
		return string.format('%s - %s', entry1, entry2)
	end)
	:gsub('<hr>', '{separator}')
	:gsub('<.->', '')
	--:gsub('^\n\n', '\n')
	:gsub('\n%s+\n', '\n\n')
	--:gsub('  \n', '\n\n')
	:gsub('{{< (%w+) `(.-)` >}}', function(shortcode, text)
		return docfuncs.renderInfoBlock(shortcode, text)
	end)
	:gsub('```(%w+)\n(.-)```', function(lang, text)
		return docfuncs.renderCodeBlock(text)
	end)
	:gsub('```\n(.-)\n```', function(text)
		return docfuncs.renderCodeBlock(text)
	end)
	:gsub('`[^\n].-`', lunacolors.cyan)
	:gsub('#+ (.-\n)', function(heading) return lunacolors.blue(lunacolors.bold('→ ' .. heading)) end)
	:gsub('%*%*(.-)%*%*', lunacolors.bold)
end

commander.register('doc', function(args, sinks)
	local moddocPath = hilbish.dataDir .. '/docs/'
	local stat = pcall(fs.stat, '.git/refs/heads/extended-job-api')
	if stat then
		-- hilbish git
		moddocPath = './docs/'
	end

	local modules = table.map(fs.readdir(moddocPath), function(f)
		return lunacolors.underline(lunacolors.blue(string.gsub(f, '.md', '')))
	end)
	local doc = [[
Welcome to Hilbish's documentation viewer! Here you can find
documentation for builtin functions and other things related
to Hilbish.

Usage: doc <section> [subdoc]
Available sections: ]] .. table.concat(modules, ', ')
	local f
	local function handleYamlInfo(d)
		local vals = {}
		local docs = d

		local valsStr = docs:match '^%-%-%-\n.-\n%-%-%-'
		if valsStr then
			docs = docs:sub(valsStr:len() + 2, #docs)
			local pre = docs:sub(1, 1)
			if pre == '\n' then
				docs = docs:sub(2)
			end

			-- parse vals
			local lines = string.split(valsStr, '\n')
			for _, line in ipairs(lines) do
				local key = line:match '(%w+): '
				local val = line:match '^%w+: (.-)$'

				if key then
					vals[key] = val
				end
			end
		end

		--docs = docs:sub(1, #docs - 1)
		return docs, vals
	end

	if #args > 0 then
		local mod = args[1]

		f = io.open(moddocPath .. mod .. '.md', 'rb')
		local funcdocs = nil
		local subdocName = args[2]
		if not f then
			moddocPath = moddocPath .. mod .. '/'
			if not subdocName then
				subdocName = '_index'
			end
			f = io.open(moddocPath .. subdocName .. '.md', 'rb')
			local oldmoddocPath = moddocPath
			if not f then
				moddocPath = moddocPath .. subdocName:match '%w+' .. '/'
				f = io.open(moddocPath .. subdocName .. '.md', 'rb')
			end
			if not f then
				moddocPath = oldmoddocPath .. subdocName .. '/'
				subdocName = args[3] or '_index'
				f = io.open(moddocPath .. subdocName .. '.md', 'rb')
			end
			if not f then
				sinks.out:writeln('No documentation found for ' .. mod .. '.')
				return 1
			end
		end

	end

	local moddocs = table.filter(fs.readdir(moddocPath), function(f) return f ~= '_index.md' and f ~= 'index.md' end)
	local subdocs = table.map(moddocs, function(fname)
		return lunacolors.underline(lunacolors.blue(string.gsub(fname, '.md', '')))
	end)

	local gh = Greenhouse(sinks.out)
	function gh:resize()
		local size = terminal.size()
		self.region = {
			width = size.width,
			height = size.height - 1
		}
	end
	gh:resize()

	function gh:render()
		local workingPage = self.pages[self.curPage]
		local offset = self.offset
		if self.isSpecial then
			offset = self.specialOffset
			workingPage = self.specialPage
		end
		local size = terminal.size()

		self.sink:write(ansikit.getCSI(size.height - 1 .. ';1', 'H'))
		self.sink:write(ansikit.getCSI(0, 'J'))
		if not self.isSpecial then
			if args[1] == 'api' then
				self.sink:writeln(workingPage.title)
				self.sink:write(lunacolors.format(string.format('{grayBg} ↳ {white}{italic}%s {reset}', workingPage.description or 'No description.')))
			else
				self.sink:write(lunacolors.reset(string.format('Viewing doc page %s', moddocPath)))
			end
		end
	end
	local backtickOccurence = 0
	local function formatDocText(d)
		return transformHTMLandMD(d)
		--[[
		return lunacolors.format(d:gsub('`(.-)`', function(t)
			return docfuncs.renderCodeBlock(t)
		end):gsub('\n#+.-\n', function(t)
			local signature = t:gsub('<.->(.-)</.->', '{underline}%1'):gsub('\\', '<')
			return '{bold}{yellow}' .. signature .. '{reset}'
		end))
		]]--
	end


	local doc, vals = handleYamlInfo(#args == 0 and doc or formatDocText(f:read '*a'))
	if #moddocs ~= 0 and f then
		doc = doc .. '\nSubdocs: ' .. table.concat(subdocs, ', ') .. '\n\n'
	end
	if f then f:close() end

	local page = Page(vals.title, doc)
	page.description = vals.description
	gh:addPage(page)

	-- add subdoc pages
	for _, sdName in ipairs(moddocs) do
		local sdFile = fs.join(sdName, '_index.md')
		if sdName:match '.md$' then
			sdFile = sdName
		end

		local f = io.open(moddocPath .. sdFile, 'rb')
		local doc, vals = handleYamlInfo(formatDocText(f:read '*a'))
		local page = Page(vals.title or sdName, doc)
		page.description = vals.description
		gh:addPage(page)
	end
	ansikit.hideCursor()
	gh:initUi()
end)
