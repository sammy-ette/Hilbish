--- page object for greenhouse
-- A page is a document/view that can be displayed in the Greenhouse pager.
local Object = require 'nature.object'

local Page = Object:extend()

--- Creates a new Page with the given title and text content.
--- @param title string The title of the page.
--- @param text string The text content of the page. Lines are split by newlines.
function Page:new(title, text)
	self:setText(text)
	self.title = title or 'Page'
	self.lazy = false
	self.loaded = true
	self.children = {}
end


--- Sets the text content of the page. The text is split into lines by newlines.
--- @param text string The new text content for the page.
function Page:setText(text)
	self.lines = string.split(text, '\n')
end

--- Sets or updates the title of the page.
--- @param title string The new title for the page.
function Page:setTitle(title)
	self.title = title
end

--- Marks the page as lazy-loaded with a dynamic initializer function.
--- This is used for pages that should load their content on demand.
--- The initializer function will be called when the page needs to be loaded.
--- @param initializer function A function that will be called to initialize the page content.
function Page:dynamic(initializer)
	self.initializer = initializer
	self.lazy = true
	self.loaded = false
end

--- Initializes a lazy-loaded page by calling its initializer function.
function Page:initialize()
	self.initializer()
	self.loaded = true
end

return Page
