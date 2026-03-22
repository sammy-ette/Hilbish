local commander = require 'commander'

commander.register('exec', function(args)
	if #args == 0 then
		return
	end
	hilbish.exec(args[1])
end)
