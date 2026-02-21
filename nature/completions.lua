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
		if ok then
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
