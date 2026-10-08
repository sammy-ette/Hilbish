local editor = require 'nature.editor'
local optionValues = {}

local optionSpecs = {
	autocd = {
		default = false,
		module = 'nature.opts.autocd'
	},
	history = {
		default = true,
		module = 'nature.opts.history'
	},
	hinter = {
		default = true,
		apply = function(value)
			hilbish.editor:setHinter(value and editor.defaultHinter or nil)
		end
	},
	greeting = {
		default = string.format([[Welcome to {magenta}Hilbish{reset}, {cyan}%s{reset}.
The nice lil shell for {blue}Lua{reset} fanatics!
]], hilbish.user),
		module = 'nature.opts.greeting'
	},
	motd = {
		default = true,
		module = 'nature.opts.motd'
	},
	fuzzy = {
		default = false
	},
	notifyJobFinish = {
		default = true,
		module = 'nature.opts.notifyJobFinish'
	},
	crimmas = {
		default = true,
		module = 'nature.opts.crimmas'
	},
	processorSkipList = {
		default = {}
	}
}

local function setOpt(name, value)
	optionValues[name] = value
	local spec = optionSpecs[name]
	if spec and spec.apply then
		spec.apply(value)
	end
end

hilbish.opts = setmetatable({}, {
	__index = optionValues,
	__newindex = function(_, name, value)
		setOpt(name, value)
	end,
	__pairs = function()
		return next, optionValues, nil
	end
})

for name, spec in pairs(optionSpecs) do
	optionValues[name] = spec.default
	if spec.module then
		pcall(require, spec.module)
	end
	if spec.apply then
		spec.apply(spec.default)
	end
end
