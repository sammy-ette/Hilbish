local lunacolors = require 'lunacolors'

local styles = {
	'bold', 'dim', 'italic', 'underline', 'invert',
	'strikethrough',
	'red', 'green', 'yellow', 'blue', 'magenta', 'cyan',
	'white', 'blackBg', 'redBg', 'greenBg', 'yellowBg',
	'blueBg', 'magentaBg', 'cyanBg', 'whiteBg'
}

for _, k in ipairs(styles) do
	if k == 'red' or k == 'redBg' then io.write '\n' end
	io.write(lunacolors[k](k) .. '  ')
end

io.write '\n'
