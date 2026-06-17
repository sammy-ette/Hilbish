local vim = {}
hilbish.vim = {
	registers = {}
}

setmetatable(hilbish.vim.registers, {
	__newindex = function(_, k, v)
		hilbish.editor:setVimRegister(k, v)
	end,
	__index = function(_, k)
		return hilbish.editor:getVimRegister(k)
	end
})

setmetatable(hilbish.vim, {
	__index = function(_, k)
		if k == 'mode' then return hilbish.vimMode end
	end
})
