-- @module hilbish.processors
-- command processing before execution
-- The processors interface manages command processors, which are functions that
-- can transform a command string before it is executed. Processors run in order
-- of priority, lowest number first.

---@diagnostic disable-next-line: missing-fields
hilbish.processors = {
	list = {},
	sorted = {}
}

--- Registers a command processor. A processor is a table with at minimum a
--- `name` and a `func` field. The `func` receives the command string and may
--- return a table with any of: `command` (the new command string), `continue`
--- (whether to abort execution if false), and [`modifiers`](../features/modifiers).
--- @param processor table A table with `name` (string), `func` (function), and optional `priority` (number, default 0).
--- @example
--- hilbish.processors.add({
--- 	name = 'my-processor',
--- 	priority = 0,
--- 	func = function(cmd)
--- 		-- do something with cmd
--- 	end
--- })
--- @example
function hilbish.processors.add(processor)
	if not processor.name then
		error 'processor is missing name'
	end

	if not processor.func then
		error 'processor is missing function'
	end

	processor.priority = processor.priority or 0

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

--- Runs all registered processors against the provided command in priority order.
--- @param command string The command string to process.
--- @param opts? table
--- @tparam opts? skip? table<string> A list of processor names to skip.
--- @return table result The (possibly modified) result of running the processors.
--- @treturn result command string The processed command string.
--- @treturn result continue boolean Whether execution should proceed.
--- @treturn result modifiers table Any modifier flags set by processors.
function hilbish.processors.execute(command, opts)
	opts = opts or {}
	opts.skip = opts.skip or {}

	local continue = true
	local modifiers = {}
	for _, processor in ipairs(hilbish.processors.list) do
		if not contains(opts.skip, processor.name) then
			local processed = processor.func(command)
			if processed then
				if processed.command then command = processed.command end
				if processed.modifiers then
					for k, v in pairs(processed.modifiers) do
						modifiers[k] = v
					end
				end
				if processed.continue == false then
					continue = false
					break
				end
			end
		end
	end

	return {
		command = command,
		continue = continue,
		modifiers = modifiers
	}
end
