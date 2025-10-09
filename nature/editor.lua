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
		local args = {...}
		if args[1] == hilbish.editor then
			table.remove(args, 1)
		end
        return editor[key](editor, table.unpack(args))
    end
end

setmetatable(hilbish.editor, editorMt)
