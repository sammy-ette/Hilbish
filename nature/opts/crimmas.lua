local lunacolors = require 'lunacolors'

bait.catch('hilbish.init', function()
	
	if os.date '%m' == '12' and hilbish.interactive and hilbish.opts.crimmas then
		local crimmas = math.random(1, 31)
		if crimmas >= 25 and crimmas <= 29 then
			print(lunacolors.format '🎄 {green}Merry {red}Christmas{reset} from your {green}favourite{reset} shell {red}(right?){reset} 🌺')
		end
	end
end)
