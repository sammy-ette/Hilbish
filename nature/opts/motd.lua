local bait = require 'bait'
local lunacolors = require 'lunacolors'

hilbish.motd = [[
{magenta}Hilbish{reset} blooms in the {blue}midnight.{reset}
]]

bait.catch('hilbish.init', function()
	if hilbish.interactive and hilbish.opts.motd then
		print(lunacolors.format(hilbish.motd))
	end
end)
