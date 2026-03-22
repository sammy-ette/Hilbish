local commander = require 'commander'

commander.register('fg', function(_, sinks)
	local job = hilbish.jobs.last()
	if not job then
		sinks.out:writeln 'fg: no last job'
		return 1
	end

	local err = job:foreground() -- waits for job; blocks
	if err then
		sinks.out:writeln('fg: ' .. err)
		return 2
	end
end)
