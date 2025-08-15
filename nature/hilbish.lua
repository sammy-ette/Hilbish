-- @module hilbish
local bait = require 'bait'
local fs = require 'fs'
local readline = require 'readline'
local snail = require 'snail'

hilbish.snail = snail.new()
hilbish.snail:run 'true' -- to "initialize" snail
bait.catch('hilbish.cd', function(path)
	hilbish.snail:dir(path)
end)

local function abbrevHome(path)
	if path:sub(1, hilbish.home:len()) == hilbish.home then
		return fs.join('~', path:sub(hilbish.home:len() + 1))
	end

	return path
end

local function expandHome(path)
	if path:sub(1, 1) == '~' then
		return fs.join(hilbish.home, path:sub(2))
	end

	return path
end

local function fmtPrompt(p)
	return p:gsub('%%(%w)', function(c)
		if c == 'd' then
			return abbrevHome(hilbish.cwd())
		elseif c == 'D' then
			return fs.basename(abbrevHome(hilbish.cwd()))
		elseif c == 'u' then
			return hilbish.user
		elseif c == 'h' then
			return hilbish.host
		end
	end)
end

--- prompt(str, typ)
--- Changes the shell prompt to the provided string.
--- There are a few verbs that can be used in the prompt text.
--- These will be formatted and replaced with the appropriate values.
--- `%d` - Current working directory
--- `%D` - Basename of working directory ()
--- `%u` - Name of current user
--- `%h` - Hostname of device
--- #param str string
--- #param typ? string Type of prompt, being left or right. Left by default.
--- #example
--- -- the default hilbish prompt without color
--- hilbish.prompt '%u %d ∆'
--- -- or something of old:
--- hilbish.prompt '%u@%h :%d $'
--- -- prompt: user@hostname: ~/directory $
--- #example
-- @param p string
-- @param typ string Type of prompt, either left or right
function hilbish.prompt(p, typ)
	if type(p) ~= 'string' then
		error('expected #1 to be string, got ' .. type(p))
	end

	if not typ or typ == 'left' then
		hilbish.editor:prompt(fmtPrompt(p))
		if not hilbish.running then
			hilbish.editor:refreshPrompt()
		end
	elseif typ == 'right' then
		hilbish.editor:rightPrompt(fmtPrompt(p))
		if not hilbish.running then
			hilbish.editor:refreshPrompt()
		end
	else
		error('expected prompt type to be right or left, got ' .. tostring(typ))
	end
end

local pathSep = ':'
if hilbish.os.family == 'windows' then
	pathSep = ';'
end

local function appendPath(path)
	os.setenv('PATH', os.getenv 'PATH' .. pathSep .. expandHome(path))
end

--- appendPath(path)
--- Appends the provided dir to the command path (`$PATH`)
--- @param path string|table Directory (or directories) to append to path
--- #example
--- hilbish.appendPath '~/go/bin'
--- -- Will add ~/go/bin to the command path.
--- 
--- -- Or do multiple:
--- hilbish.appendPath {
--- 	'~/go/bin',
--- 	'~/.local/bin'
--- }
--- #example
function hilbish.appendPath(path)
	if type(path) == 'table' then
		for _, p in ipairs(path) do
			appendPath(p)
		end
	elseif type(path) == 'string' then
		appendPath(path)
	else
		error('bad argument to appendPath (expected string or table, got ' .. type(path) .. ')')
	end
end

local function prependPath(path)
	print('prepending', path, expandHome(path))
	os.setenv('PATH', expandHome(path) .. pathSep .. os.getenv 'PATH')
end

--- prependPath(path)
--- Prepends the provided dir to the command path (`$PATH`)
--- @param path string|table Directory (or directories) to append to path
--- #example
--- hilbish.prependPath '~/go/bin'
--- -- Will add ~/go/bin to the command path.
--- 
--- -- Or do multiple:
--- hilbish.prependPath {
--- 	'~/go/bin',
--- 	'~/.local/bin'
--- }
--- #example
function hilbish.prependPath(path)
	if type(path) == 'table' then
		for _, p in ipairs(path) do
			prependPath(p)
		end
	elseif type(path) == 'string' then
		prependPath(path)
	else
		error('bad argument to prependPath (expected string or table, got ' .. type(path) .. ')')
	end
end

--- read(prompt) -> input (string)
--- Read input from the user, using Hilbish's line editor/input reader.
--- This is a separate instance from the one Hilbish actually uses.
--- Returns `input`, will be nil if Ctrl-D is pressed, or an error occurs.
-- @param prompt? string Text to print before input, can be empty.
-- @returns string|nil
function hilbish.read(prompt)
	prompt = prompt or ''
	if type(prompt) ~= 'string' then
		error 'expected #1 to be a string'
	end

	local rl = readline.new()
	rl:prompt(prompt)

	return rl:read()
end

--- Runs `cmd` in Hilbish's shell script interpreter.
--- The `streams` parameter specifies the output and input streams the command should use.
--- For example, to write command output to a sink.
--- As a table, the caller can directly specify the standard output, error, and input
--- streams of the command with the table keys `out`, `err`, and `input` respectively.
--- As a boolean, it specifies whether the command should use standard output or return its output streams.
--- #example
--- -- This code is the same as `ls -l | wc -l`
--- local fs = require 'fs'
--- local pr, pw = fs.pipe()
--- hilbish.run('ls -l', {
--- 	stdout = pw,
--- 	stderr = pw,
--- })
--- pw:close()
--- hilbish.run('wc -l', {
--- 	stdin = pr
--- })
--- #example
-- @param cmd string
-- @param streams table|boolean
-- @returns number, string, string
function hilbish.run(cmd, streams)
	local sinks = {}

	if type(streams) == 'boolean' then
		if not streams then
			sinks = {
				out = hilbish.sink.new(),
				err = hilbish.sink.new(),
				input = io.stdin
			}
		end
	elseif type(streams) == 'table' then
		sinks = streams
	end

	local out = hilbish.snail:run(cmd, {sinks = sinks})
	local returns = {out.exitCode}

	if type(streams) == 'boolean' and not streams then
		table.insert(returns, sinks.out:readAll())
		table.insert(returns, sinks.err:readAll())
	end

	return table.unpack(returns)
end

--- Sets the execution/runner mode for interactive Hilbish.
--- **NOTE: This function is deprecated and will be removed in 3.0**
--- Use `hilbish.runner.setCurrent` instead.
--- This determines whether Hilbish wll try to run input as Lua
--- and/or sh or only do one of either.
--- Accepted values for mode are hybrid (the default), hybridRev (sh first then Lua),
--- sh, and lua. It also accepts a function, to which if it is passed one
--- will call it to execute user input instead.
--- Read [about runner mode](../features/runner-mode) for more information.
-- @param mode string|function
function hilbish.runnerMode(mode)
	if type(mode) == 'string' then
		hilbish.runner.setCurrent(mode)
	elseif type(mode) == 'function' then
		hilbish.runner.set('_', {
			run = mode
		})
		hilbish.runner.setCurrent '_'
	else
		error('expected runner mode type to be either string or function, got', type(mode))
	end
end

local multilinePrompt = '~> '

--- Changes the text prompt when Hilbish asks for more input.
--- This will show up when text is incomplete, like a missing quote
--- @param str string
--- #example
--- --[[
--- imagine this is your text input:
--- user ~ ∆ echo "hey
--- but there's a missing quote! hilbish will now prompt you so the terminal
--- will look like:
--- user ~ ∆ echo "hey
--- --> ...!"

--- so then you get
--- user ~ ∆ echo "hey
--- --> ...!"
--- hey ...!
--- ]]--
--- hilbish.multiprompt '-->'
--- #example
function hilbish.multiprompt(str)
	if str == nil then
		return multilinePrompt
	end

	assert(type(str) == 'string', 'expected multiprompt to be a string, found ' .. type(str))
	multilinePrompt = str
end
