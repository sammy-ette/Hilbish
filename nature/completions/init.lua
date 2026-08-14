local fs = require 'fs'

-- explanation: this specific function gives to us info about
-- the currently running source. this includes a path to the
-- source file (info.source)
-- we will use that to automatically load all commands by reading
-- all the files in this dir and just requiring it.
local info = debug.getinfo(1)
local commandDir = fs.dir(info.source:gsub('@', ''))
if commandDir == '.' then return end

local commands = fs.readdir(commandDir)
for _, command in ipairs(commands) do
	local name = command:gsub('%.lua', '') -- chop off extension
	if name ~= 'init' then
		-- skip this file (for obvious reasons)
		require('nature.completions.' .. name)
	end
end

function hilbish.completions.handler(line, pos)
	if type(line) ~= 'string' then error '#1 must be a string' end
	if type(pos) ~= 'number' then error '#2 must be a number' end

	-- trim leading whitespace
	local ctx = line:gsub('^%s*(.-)$', '%1')
	if ctx:len() == 0 then return {}, '' end

	local res = hilbish.aliases.resolve(ctx)
	local resFields = string.split(res, ' ')
	local fields = string.split(ctx, ' ')
	if #fields > 1 and #resFields > 1 then
		fields = resFields
	end
	local query = fields[#fields]

	if query:match('^@.+') then
		local name = query:match '^@([a-zA-Z0-9]+)'
		local val = query:match '^@[a-zA-Z0-9]+=(.*)' or ''

		if name == 'dir' then
			local comps, pfx = hilbish.completions.dirs(val, val, {val})
			local compGroup = {
				items = comps,
				type = 'grid'
			}
			return {compGroup}, pfx
		end

		return {}, ''
	end

	while #fields > 0 and fields[1]:match('^@.+') do
		table.remove(fields, 1)
	end

	if #fields == 1 then
		local comps, pfx = hilbish.completions.bins(query, ctx, fields)
		local compGroup = {
			items = comps,
			type = 'grid'
		}

		return {compGroup}, pfx
	else
		local ok, compGroups, pfx = pcall(hilbish.completions.call,
		'command.' .. fields[1], query, ctx, fields)
		if ok and compGroups then
			return compGroups, pfx
		end

		local comps, pfx = hilbish.completions.files(query, ctx, fields)
		local compGroup = {
			items = comps,
			type = 'grid'
		}

		return {compGroup}, pfx
	end
end
