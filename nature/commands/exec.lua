local commander = require 'commander'
local hilbish = require 'hilbish'

commander.register('exec', function(args)
	if #args == 0 then
		return
	end
	hilbish.exec(args[1])
end)
