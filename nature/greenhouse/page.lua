-- @module greenhouse.page
local Object = require 'nature.object'

local Page = Object:extend()

function Page:new(title, text)
	self:setText(text)
	self.title = title or 'Page'
	self.lazy = false
	self.loaded = true
	self.children = {}
end


function Page:setText(text)
	self.lines = string.split(text, '\n')
end

function Page:setTitle(title)
	self.title = title
end

function Page:dynamic(initializer)
	self.initializer = initializer
	self.lazy = true
	self.loaded = false
end

function Page:initialize()
	self.initializer()
	self.loaded = true
end

return Page
