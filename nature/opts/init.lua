hilbish.opts = {}

local function setupOpt(name, default)
	hilbish.opts[name] = default
	local ok, err = pcall(require, 'nature.opts.' .. name)
end

local defaultOpts = {
	autocd = false,
	history = true,
	greeting = string.format([[Welcome to {magenta}Hilbish{reset}, {cyan}%s{reset}.
The nice lil shell for {blue}Lua{reset} fanatics!
]], hilbish.user),
	motd = true,
	fuzzy = false,
	notifyJobFinish = true,
	crimmas = true,
	tips = true,
	processorSkipList = {}
}

for optsName, default in pairs(defaultOpts) do
	setupOpt(optsName, default)
end
