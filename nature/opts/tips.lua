local bait = require 'bait'
local lunacolors = require 'lunacolors'

local postamble = [[
{yellow}These tips can be disabled with {reset}{invert} hilbish.opts.tips = false {reset}
]]

hilbish.tips = {
	'Join the discord and say hi! {blue}https://discord.gg/3PDdcQz{reset}',
	'{green}hilbish.alias{reset} interface manages shell aliases. See more detail by running {blue}doc api hilbish.alias.',
	'{green}hilbish.appendPath(\'path\'){reset} -> Appends the provided dir to the command path ($PATH)',
	'{green}hilbish.completions{reset} -> Used to control suggestions when tab completing.',
	'{green}hilbish.message{reset} -> Simple notification system which can be used by other plugins and parts of the shell to notify the user of various actions.',
	[[
{green}hilbish.opts{reset} -> Simple toggle or value options a user can set.
You may disable the startup greeting by {invert}hilbish.opts.greeting = false{reset}
]],
[[
{green}hilbish.runner{reset} -> The runner interface contains functions to
manage how Hilbish interprets interactive input. The default runners can run
shell script and Lua code!
]],
[[
Add Lua-written commands with the commander module!
Check the command {blue}doc api commander{reset} or the web docs:
https://hilbish.sammyette.party/docs/api/commander/
]]
}

bait.catch('hilbish.init', function()
	if hilbish.interactive and hilbish.opts.tips then
		local idx = math.random(1, #hilbish.tips)
		print(lunacolors.format('{yellow}🛈 Tip:{reset} ' .. hilbish.tips[idx] .. '\n' .. postamble))
	end
end)
