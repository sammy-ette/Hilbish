local fs = require 'fs'

hilbish.processors.add {
	name = 'hilbish.autocd',
	func = function(path)
		if hilbish.opts.autocd then
			local ok, stat = pcall(fs.stat, path)
			if ok and stat.isDir then
				local oldPath = hilbish.cwd()

				local absPath = fs.abs(path)
				fs.cd(path)

				bait.throw('cd', path, oldPath)
				bait.throw('hilbish.cd', absPath, oldPath)

			end
			return {
				continue = not ok
			}
		else
			return {
				continue = true
			}
		end
	end
}
