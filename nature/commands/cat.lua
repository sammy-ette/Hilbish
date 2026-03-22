local commander = require 'commander'
local fs = require 'fs'

commander.register('cat', function(args, sinks)
	local exit = 0

	if #args == 0 then
		sinks.out:writeln [[
usage: cat [file]...]]
	end

	local chunkSize = 2^13 -- 8K buffer size

	for _, fName in ipairs(args) do
		local f = io.open(fName)
		if f == nil then
			exit = 1
			sinks.out:writeln(string.format('cat: %s: no such file or directory', fName))
			goto continue
		end

		while true do
			local block = f:read(chunkSize)
			if not block then break end
			sinks.out:write(block)
		end
		::continue::
	end
	io.flush()
	return exit
end)
