local commander = require 'commander'

commander.register('disown', function(args, sinks)
	if #hilbish.jobs.all() == 0 then
		sinks.out:writeln 'disown: no current job'
		return 1
	end

	local id
	if #args < 0 then
		id = tonumber(args[1])
		if not id then
			sinks.out:writeln 'disown: invalid id for job'
			return 1
		end
	else
		id = hilbish.jobs.last().id
	end

	local ok = pcall(hilbish.jobs.disown, id)
	if not ok then
		sinks.out:writeln 'disown: job does not exist'
		return 2
	end
end)
