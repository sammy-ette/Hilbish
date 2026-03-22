local bait = require 'bait'
local lunacolors = require 'lunacolors'

bait.catch('job.done', function(job)
	if not hilbish.opts.notifyJobFinish then return end
	local notifText = string.format(lunacolors.format [[
Background job with ID#%d has exited (PID %d).
Command string: {bold}{yellow}%s{reset}]], job.id, job.pid, job.cmd)

	if job.stdout ~= '' then
		notifText = notifText .. '\n\nStandard output:\n' .. job.stdout
	end
	if job.stderr ~= '' then
		notifText = notifText .. '\n\nStandard error:\n' .. job.stderr
	end

	hilbish.messages.send {
		channel = 'jobNotify',
		title = string.format('Job ID#%d Exited', job.id),
		summary = string.format(lunacolors.format 'Background job with command {bold}{yellow}%s{reset} has finished running!', job.cmd),
		text = notifText
	}
end)
