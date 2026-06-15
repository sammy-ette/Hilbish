local readline = require 'readline'

local editor = readline.new()
local editorMt = {}

hilbish.editor = {}

local function contains(search, needle)
	for _, p in ipairs(search) do
		if p == needle then
			return true
		end
	end

	return false
end

function editorMt.__index(_, key)
    if contains({'deleteByAmount', 'getVimRegister', 'getLine', 'insert', 'readChar', 'setVimRegister'}, key) then
		--editor:log 'The calling method of this function has changed. Please use the colon to call this hilbish.editor function.'
    end

	return function(...)
        return editor[key](editor, ...)
    end
end

setmetatable(hilbish.editor, editorMt)
