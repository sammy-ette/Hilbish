-- Prelude initializes everything else for our shell
local _ = require 'succulent' -- Function additions
local bait = require 'bait'
local fs = require 'fs'

package.path = package.path .. ';' .. hilbish.dataDir .. '/?/init.lua'
.. ';' .. hilbish.dataDir .. '/?/?.lua' .. ";" .. hilbish.dataDir .. '/?.lua'

hilbish.module.paths = '?.so;?/?.so;'
.. hilbish.userDir.data .. 'hilbish/libs/?/?.so'
.. ";" .. hilbish.userDir.data .. 'hilbish/libs/?.so'

table.insert(package.searchers, function(module)
	local path = package.searchpath(module, hilbish.module.paths)
	if not path then return nil end

	-- it didnt work normally, idk
	return function() return hilbish.module.load(path) end, path
end)

require 'nature.editor'
require 'nature.hilbish'
require 'nature.processors'

require 'nature.commands'
require 'nature.completions'
require 'nature.opts'
require 'nature.vim'
require 'nature.runner'
require 'nature.hummingbird'
require 'nature.env'
require 'nature.abbr'

local shlvl = tonumber(os.getenv 'SHLVL')
if shlvl ~= nil then
	os.setenv('SHLVL', tostring(shlvl + 1))
else
	os.setenv('SHLVL', '0')
end

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

	local _, err = pcall(dofile, path)
	if err then
		print(err)
		print 'An error has occured while loading your config!\n'
		hilbish.prompt '& '
	else
		bait.throw 'hilbish.init'
	end
end

local _, err = pcall(fs.stat, hilbish.confFile)
if err and tostring(err):match 'no such file' and hilbish.confFile == fs.join(hilbish.defaultConfDir, 'init.lua') then
	-- Run config from current directory (assuming this is Hilbish's git)
	local _, err = pcall(fs.stat, '.hilbishrc.lua')
	local confpath = '.hilbishrc.lua'

	if err then
		-- If it wasnt found go to system sample config
		confpath = fs.join(hilbish.dataDir, confpath)
		local _, err = pcall(fs.stat, confpath)
		if err then
			print('could not find .hilbishrc.lua or ' .. confpath)
			return
		end
	end

	runConfig(confpath)
else
	runConfig(hilbish.confFile)
end

-- TODO: hilbish.exit function, stop jobs and timers.
local function exit(code)
	os.exit(code)
end

while hilbish.interactive do
	hilbish.running = false

	local ok, res = pcall(function() return hilbish.editor:read() end)
	if not ok and tostring(res):lower():match 'eof' then
		bait.throw 'hilbish.exit'
		exit(0)
	end
	if not ok then
		if tostring(res):lower():match 'ctrl%+c' then
			print '^C'
			bait.throw 'hilbish.cancel'
		else
			error(res)
			io.read()
		end
		goto continue
	end
	--- @type string
	local input = res

	local priv = false
	if res:sub(1, 1) == ' ' then
		priv = true
	end
	input = input:gsub('%s+', '')

	if input:len() == 0 then
		hilbish.running = true
		bait.throw('command.exit', 0 )
		goto continue
	end

	hilbish.running = true
	hilbish.runner.run(input, priv)

	::continue::
end
