local hintHistory = {}
local hintHistorySize = -1

local function refreshHintHistory()
	local size = hilbish.history.size()
	if size == hintHistorySize then return end

	local byCommand = {}
	for index = size - 1, 0, -1 do
		local command = hilbish.history.get(index)
		if not command:find('\n', 1, true) then
			local item = byCommand[command]
			if item then
				item.count = item.count + 1
			else
				byCommand[command] = {
					command = command,
					count = 1,
					last = index
				}
			end
		end
	end

	hintHistory = {}
	for _, item in pairs(byCommand) do
		table.insert(hintHistory, item)
	end

	table.sort(hintHistory, function(a, b)
		if a.count ~= b.count then return a.count > b.count end
		return a.last > b.last
	end)

	hintHistorySize = size
end

function hilbish.hinter(line, pos)
	if not hilbish.opts.hinter or not hilbish.opts.history then return '' end
	if line == '' or pos ~= #line then return '' end

	refreshHintHistory()

	for _, item in ipairs(hintHistory) do
		if item.command:sub(1, #line) == line and item.command ~= line then
			return item.command:sub(#line + 1)
		end
	end

	return ''
end
