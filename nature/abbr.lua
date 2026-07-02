-- @module hilbish.abbr
-- @since 3.0
-- command line abbreviations
-- The abbr module manages Hilbish abbreviations. These are words that can be replaced
-- with longer command line strings when entered.
-- As an example, `git push` can be abbreviated to `gp`. When the user types
-- `gp` into the command line, after hitting space or enter, it will expand to `git push`.
-- Abbreviations can be used as an alternative to aliases. They are saved entirely in the history
-- Instead of the aliased form of the same command.
local bait = require 'bait'
local hilbish = require 'hilbish'
hilbish.abbr = {
	all = {}
}

--- Adds an abbreviation.
--- When the user types `abbr` followed by space or enter, it is replaced with `expanded`.
--- If `expanded` is a function, it is called and its return value is used instead.
--- @param abbr string The abbreviation to define.
--- @param expanded string|function The string (or function returning a string) it expands to.
--- @param opts? table
--- @tparam opts? anywhere boolean If the abbreviation should expand anywhere in the line instead of only at the start.
--- @see hilbish.aliases.add
--- @example
--- hilbish.abbr.add('gp', 'git push')
--- hilbish.abbr.add('date', function() return os.date('%Y-%m-%d') end)
--- hilbish.abbr.add('--help', '--help | greenhouse', { anywhere = true })
--- @example
function hilbish.abbr.add(abbr, expanded, opts)
	opts = opts or {}
	opts.abbr = abbr
	opts.expand = expanded
	hilbish.abbr.all[abbr] = opts
end

--- Removes the named `abbr`.
--- @param abbr string
function hilbish.abbr.remove(abbr)
	hilbish.abbr.all[abbr] = nil
end

bait.catch('hilbish.rawInput', function(c)
	-- 0x0d == enter
	if c == ' ' or c == string.char(0x0d) then
		-- check if the last "word" was a valid abbreviation
		local line = hilbish.editor:getLine()
		local lineSplits = string.split(line, ' ')
		local thisAbbr = hilbish.abbr.all[lineSplits[#lineSplits]]

		if thisAbbr and (#lineSplits == 1 or thisAbbr.anywhere == true) then
			hilbish.editor:deleteByAmount(-lineSplits[#lineSplits]:len())
			if type(thisAbbr.expand) == 'string' then
				hilbish.editor:insert(thisAbbr.expand)
			elseif type(thisAbbr.expand) == 'function' then
				local expandRet = thisAbbr.expand()
				if type(expandRet) ~= 'string' then
					print(string.format('abbr %s has an expand function that did not return a string. instead it returned: %s', thisAbbr.abbr, expandRet))
					return
				end
				hilbish.editor:insert(expandRet)
			end
		end
	end
end)
