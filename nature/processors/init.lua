-- @module hilbish.processors

hilbish.processors = {
	list = {},
	sorted = {}
}

function hilbish.processors.add(processor)
	if not processor.name then
		error 'processor is missing name'
	end

	if not processor.func then
		error 'processor is missing function'
	end

	table.insert(hilbish.processors.list, processor)
	table.sort(hilbish.processors.list, function(a, b) return a.priority < b.priority end)
end

local function contains(search, needle)
	for _, p in ipairs(search) do
		if p == needle then
			return true
		end
	end

	return false
end

--- Run all command processors, in order by priority.
--- It returns the processed command (which may be the same as the passed command)
--- and a boolean which states whether to proceed with command execution.
function hilbish.processors.execute(command, opts)
	opts = opts or {}
	opts.skip = opts.skip or {}

	local continue = true
	local history
	for _, processor in ipairs(hilbish.processors.list) do
		if not contains(opts.skip, processor.name) then
			local processed = processor.func(command)
			if processed then
				if processed.history ~= nil then history = processed.history end
				if processed.command then command = processed.command end
				if not processed.continue then
					continue = false
					break
				end
			end
		end
	end

	return {
		command = command,
		continue = continue,
		history = history
	}
end