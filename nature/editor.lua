local readline = require 'readline'
local bait = require 'bait'

hilbish.editor = readline.new()

local defaultHistPath = hilbish.userDir.data .. '/hilbish/.hilbish-history'

function hilbish.hinter(line, pos)
	return ''
end

function hilbish.highlighter(line)
	return line
end

hilbish.editor:setHinter(function(line, pos)
	return hilbish.hinter(line, pos)
end)

hilbish.editor:setHighlighter(function(line)
	return hilbish.highlighter(line)
end)

hilbish.editor:setCompleter(function(line, pos)
	return hilbish.completions.handler(line, pos)
end)

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
