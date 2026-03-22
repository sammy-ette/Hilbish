local commander = require 'commander'

commander.register('bg', function(_, sinks)
	local job = hilbish.jobs.last()
	if not job then
		sinks.out:writeln 'bg: no last job'
		return 1
	end

	local err = job:background()
	if err then
		sinks.out:writeln('bg: ' .. err)
		return 2
	end
end)
