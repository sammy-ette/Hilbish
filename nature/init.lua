-- Prelude initializes everything else for our shell
local _ = require 'succulent' -- Function additions
local bait = require 'bait'
local fs = require 'fs'

hilbish.initialized = false

local oldOsExit = os.exit
---@diagnostic disable-next-line: duplicate-set-field
function os.exit(code)
	hilbish.jobs.stopAll()
	if not hilbish.interactive then
		hilbish.timers.wait()
	end
	oldOsExit(code or 0)
end

package.path = package.path .. ';' .. hilbish.dataDir .. '/?/init.lua'
.. ';' .. hilbish.dataDir .. '/?/?.lua' .. ";" .. hilbish.dataDir .. '/?.lua'

if not hilbish.midnightEdition then
	hilbish.module.paths = '?.so;?/?.so;'
	.. hilbish.userDir.data .. 'hilbish/libs/?/?.so'
	.. ";" .. hilbish.userDir.data .. 'hilbish/libs/?.so'

	table.insert(package.searchers, function(module)
		local path = package.searchpath(module, hilbish.module.paths)
		if not path then return nil end

		-- it didnt work normally, idk
		return function() return hilbish.module.load(path) end, path
	end)
else
---@diagnostic disable-next-line: undefined-global
	pcall = unsafe_pcall
end

require 'nature.editor'
require 'nature.aliases'
require 'nature.hilbish'

require 'nature.processors'
require 'nature.processors.wildcardWarn'

require 'nature.commands'
require 'nature.completions'
require 'nature.opts'
require 'nature.vim'
require 'nature.runner'
require 'nature.hummingbird'
require 'nature.env'
require 'nature.abbr'

require 'nature.paperbush'

local shlvl = tonumber(os.getenv 'SHLVL')
if shlvl ~= nil then
	os.setenv('SHLVL', tostring(shlvl + 1))
else
	os.setenv('SHLVL', '0')
end

--[[
do
	local startSearchPath = hilbish.userDir.data .. '/hilbish/start/?/init.lua;'
	.. hilbish.userDir.data .. '/hilbish/start/?.lua'

	local ok, modules = pcall(fs.readdir, hilbish.userDir.data .. '/hilbish/start/')
	if ok then
		for _, module in ipairs(modules) do
			local entry = package.searchpath(module, startSearchPath)
			if entry then
				dofile(entry)
			end
		end
	end

	package.path = package.path .. ';' .. startSearchPath
end
]]--

bait.catch('error', function(event, handler, err)
	print(string.format('Encountered an error in %s handler\n%s', event, err:sub(8)))
end)

bait.catch('command.not-found', function(cmd)
	print(string.format('hilbish: %s not found', cmd))
end)

bait.catch('command.not-executable', function(cmd)
	print(string.format('hilbish: %s: not executable', cmd))
end)

local function runConfig(path)
	if not hilbish.interactive then return end

	local ok, err = pcall(dofile, path)
	if not ok then
		print(err)
		print 'An error has occurred while loading your config!\n'
		hilbish.prompt '& '
	else
		bait.throw 'hilbish.init'
	end
end

local ok, ret = pcall(fs.stat, hilbish.confFile)
if not ok and tostring(ret):match 'no such file' and hilbish.confFile == fs.join(hilbish.defaultConfDir, 'init.lua') then
	-- Run config from current directory (assuming this is Hilbish's git)
	local ok = pcall(fs.stat, '.hilbishrc.lua')
	local confpath = '.hilbishrc.lua'

	if not ok then
		-- If it wasnt found go to system sample config
		confpath = fs.join(hilbish.dataDir, confpath)
		local ok = pcall(fs.stat, confpath)
		if not ok then
			print('could not find .hilbishrc.lua or ' .. confpath)
			return
		end
	end

	runConfig(confpath)
else
	runConfig(hilbish.confFile)
end

-- Input piped to Hilbish (not interactive, and not because a file or command
-- was passed). Run each line, then exit.
if not hilbish.interactive and not args[0] and hilbish.command == '' then
	for line in io.lines() do
		hilbish.runner.run(line, true)
	end
	os.exit(0)
end

if hilbish.command ~= "" then
	hilbish.runner.run(hilbish.command, true)
end

if args[0] then
	local ok, err = pcall(dofile, args[0])
	if not ok then
		io.stderr:write(err .. '\n')
		os.exit(1)
	end
	os.exit(0)
end

hilbish.initialized = true
require 'nature.repl'
