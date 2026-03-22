hilbish.complete('command.cd', function (query, ctx, fields)
    local comps, pfx = hilbish.completions.dirs(query, ctx, fields)
	local compGroup = {
		items = comps,
		type = 'grid'
	}

	return {compGroup}, pfx
end)