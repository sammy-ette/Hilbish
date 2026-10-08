local readline = require 'readline'
local bait = require 'bait'

hilbish.editor = readline.new()

local defaultHistPath = hilbish.userDir.data .. '/hilbish/.hilbish-history'

function hilbish.highlighter(line)
	return line
end

hilbish.editor:setHighlighter(function(line)
	return hilbish.highlighter(line)
end)

hilbish.editor:setCompleter(function(line, pos)
	return hilbish.completions.handler(line, pos)
end)

-- see nature/hooks.lua for hilbish.vimMode/hilbish.vimAction/hilbish.rawInput hook docs
hilbish.editor:setViModeCallback(function(mode)
	hilbish.vimMode = mode
	bait.throw('hilbish.vimMode', mode)
end)

hilbish.editor:setViActionCallback(function(action, args)
	bait.throw('hilbish.vimAction', action, args)
end)

hilbish.editor:setRawInputCallback(function(input)
	bait.throw('hilbish.rawInput', input)
end)

-- Returns nil if fuzzy search is off, falling back to Go's default regex searcher
hilbish.editor:setSearcher(function(needle, haystack)
	if hilbish.opts.fuzzy then
		return readline.fuzzySearch(needle, haystack)
	end
end)

local hist = readline.newHistory(defaultHistPath)
hilbish.history = hist
hilbish.editor:setHistory(hist)

function hilbish.inputMode(mode)
	if mode == 'emacs' then
		hilbish.vimMode = nil
		hilbish.editor:setInputMode('emacs')
	elseif mode == 'vim' then
		hilbish.vimMode = 'insert'
		bait.throw('hilbish.vimMode', 'insert')
		hilbish.editor:setInputMode('vim')
	else
		error('inputMode: expected vim or emacs, got ' .. mode)
	end
end

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

local function defaultHinter(line, pos)
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

return {defaultHinter = defaultHinter}
