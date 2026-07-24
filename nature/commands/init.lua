local fs = require 'fs'

-- explanation: this specific function gives to us info about
-- the currently running source. this includes a path to the
-- source file (info.source)
-- we will use that to automatically load all commands by reading
-- all the files in this dir and just requiring it.
local info = debug.getinfo(1)
local commandDir = fs.dir(info.source:gsub('@', ''))
if commandDir == '.' then return end

local commands = fs.readdir(commandDir)
for _, command in ipairs(commands) do
	local name = command:gsub('%.lua', '') -- chop off extension
	if name ~= 'init' then
		-- skip this file (for obvious reasons)
		require('nature.commands.' .. name)
	end
end
