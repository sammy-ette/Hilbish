-- @module hilbish.runner
-- The runner interface contains functions that allow the user to change
--- how Hilbish interprets interactive input.
--- Users can add and change the default runner for interactive input to any
--- language or script of their choosing. A good example is using it to
--- write commands in Fennel.
--- 
--- A runner is a table with two required functions:
--- - `run(input) -> table`: Evaluates the input and returns a result table.
--- - `validate(input) -> boolean`: Checks whether the input is complete and
--- ready to run. Return `false` to prompt the user for more input (continuation),
--- or `true` to proceed.
--- 
--- The table returned by `run` can have these fields.
--- All are optional; only set the ones relevant to the runner.
--- (So if there isn't an error, just omit `err`.)
--- 
--- - `exitCode` (number): Exit code of the command
--- - `input` (string): The text input of the user. This is used by Hilbish to append extra input, in case
--- more is requested.
--- - `err` (string): A string that represents an error from the runner.
--- This should only be set when, for example, there is a syntax error.
--- It can be set to a few special values for Hilbish to throw the right
--- hooks and have a better looking message.
--- 	- `<command>: not-found` will throw a `command.not-found` hook
--- 	based on what `<command>` is.
--- 	- `<command>: not-executable` will throw a `command.not-executable` hook.
--- - `continue` (boolean): Whether Hilbish should prompt the user for more input
--- - `newline` (boolean): Whether a newline should be added at the end of `input`.
--- 
-- Here is a simple example of a fennel runner. It falls back to
--- shell script if fennel eval has an error.
--- ```lua
--- local fennel = require 'fennel'
--- 
--- hilbish.runner.add('fennel', {
--- 	run = function(input)
--- 		local ok = pcall(fennel.eval, input)
--- 		if ok then
--- 			return { input = input }
--- 		end
--- 		return hilbish.runner.sh(input)
--- 	end,
--- 	validate = function(input)
--- 		return someMethodUsedToCheckIfFennelInputIsFinished(input)
--- 	end,
--- })
--- hilbish.runner.setCurrent('fennel')
--- ```
local bait = require 'bait'
local snail = require 'snail'
local currentRunner = 'hybrid'
local runners = {}

---@diagnostic disable-next-line: missing-fields
hilbish.runner = {}

--- Get a runner by name.
--- @param name string Name of the runner to retrieve.
--- @return table
function hilbish.runner.get(name)
	local r = runners[name]

	if not r then
		error(string.format('runner %s does not exist', name))
	end

	return r
end

--- Adds a runner to the table of available runners.
--- `runner` must be a table with both a `run` and a `validate` function.
--- @param name string Name of the runner
--- @param runner table
function hilbish.runner.add(name, runner)
	if runners[name] then
		error(string.format('runner %s already exists', name))
	end

	hilbish.runner.set(name, runner)
end

--- *Sets* a runner by name. The difference between this function and
--- add, is set will *not* check if the named runner exists.
--- The runner table must have both a `run` and a `validate` function.
--- @param name string
--- @param runner table
function hilbish.runner.set(name, runner)
	if type(name) ~= 'string' then
		error 'expected runner name to be a string'
	end

	if type(runner) ~= 'table' then
		error 'expected runner to be a table'
	end

	if not runner.run or type(runner.run) ~= 'function' then
		error('runner function in runner ' .. name .. ' missing')
	end

	if not runner.validate or type(runner.validate) ~= 'function' then
		error('validate function in runner ' .. name .. ' missing')
	end

	runners[name] = runner
end

--- Executes `cmd` with a runner.
--- If `runnerName` is not specified, it uses the default Hilbish runner.
--- @param cmd string
--- @param runnerName string?
--- @return table
function hilbish.runner.exec(cmd, runnerName)
	if not runnerName then runnerName = currentRunner end

	local r = hilbish.runner.get(runnerName)

	return r.run(cmd)
end

--- Sets Hilbish's runner mode by name.
--- @param name string
function hilbish.runner.setCurrent(name)
	hilbish.runner.get(name) -- throws if it doesnt exist.
	currentRunner = name
end

--- Returns the current runner by name.
--- @return string
function hilbish.runner.getCurrent()
	return currentRunner
end

local function finishExec(exitCode, input, priv)
	hilbish.exitCode = exitCode
	bait.throw('command.exit', exitCode, input, priv)
end

--- Prompts for continued input from the user.
--- @param prev string
--- @param newline boolean
--- @return string|nil
--- @private
function hilbish.runner.continuePrompt(prev, newline)
	local multilinePrompt = hilbish.multiprompt()
	-- the return of hilbish.read is nil when error or ctrl-d
	local cont = hilbish.read(multilinePrompt)
	if not cont then
		return
	end

	if newline then
		cont = '\n' .. cont
	end

	if cont:match '\\$' then
		cont = cont:gsub('\\$', '') .. '\n'
	end

	return prev .. cont
end

--- Runs `input` with the currently set Hilbish runner.
--- This method is how Hilbish executes commands.
--- `priv` is an optional boolean used to state if the input should be saved to history.
--- @param input string
--- @param priv? boolean
function hilbish.runner.run(input, priv)
	bait.throw('command.preprocess', input)
	local processed = hilbish.processors.execute(input, {
		skip = hilbish.opts.processorSkipList
	})
	priv = processed.history ~= nil and (not processed.history) or priv
	if not processed.continue then
		finishExec(0, '', true)
		return
	end

	local runner = hilbish.runner.get(currentRunner)
	
	::rerun::
	local command = hilbish.aliases.resolve(processed.command)
	local valid = runner.validate(processed.command)
	if not valid then
		local contInput = hilbish.runner.continuePrompt(processed.command, false)
		if contInput then
			processed.command = contInput
			goto rerun
		else
			finishExec(0, '', true)
			return
		end
	end

	bait.throw('command.preexec', processed.command, command)

	local ok, out = pcall(runner.run, processed.command)
	if not ok then
		io.stderr:write(out .. '\n')
		finishExec(124, out.input, priv)
		return
	end

	-- TODO: maybe remove this, because validate should check if
	-- input should be continued
	if out.continue then
		local contInput = hilbish.runner.continuePrompt(processed.command, out.newline)
		if contInput then
			processed.command = contInput
			goto rerun
		end
	end

	if out.err then
		local fields = string.split(out.err, ': ')
		if fields[2] == 'not-found' or fields[2] == 'not-executable' then
			bait.throw('command.' .. fields[2], fields[1])
		else
			io.stderr:write(out.err .. '\n')
		end
	end
	finishExec(out.exitCode, out.input, priv)
end

--- Runs input using Hilbish's snail instance, that is, as shell script.
--- @param input string
--- @return table
function hilbish.runner.sh(input)
	return hilbish.snail:run(input)
end

--- Evaluates `cmd` as Lua input. This is the same as using `dofile`
--- or `load`, but is appropriated for the runner interface.
--- @param input string
--- @return table
function hilbish.runner.lua(input)
	local fun, err = load(input)
	if err then
		return {
			input = input,
			exitCode = 125,
			err = err
		}
	end

	local ok
---@diagnostic disable-next-line: param-type-mismatch
	ok, err = pcall(fun)
	if not ok then
		return {
			input = input,
			exitCode = 126,
			err = err
		}
	end

	return {
		input = input,
		exitCode = 0,
	}
end

hilbish.runner.add('hybrid', {
	run = function(input)
		local cmdStr = hilbish.aliases.resolve(input)

		local res = hilbish.runner.lua(cmdStr)
		if not res.err then
			return res
		end

		return hilbish.runner.sh(input)
	end,
	validate = snail.validate
})

local function luaValidate(input)
	local f, err = load(input)
	if f then
		return true
	elseif err and err:find('<eof>') then
		return false
	else
		return true
	end
end

hilbish.runner.add('hybridRev', {
	run = function(input)
		local res = hilbish.runner.sh(input)
		if not res.err then
			return res
		end

		local cmdStr = hilbish.aliases.resolve(input)
		return hilbish.runner.lua(cmdStr)
	end,
	validate = luaValidate
})

hilbish.runner.add('lua', {
	run = function(input)
		local cmdStr = hilbish.aliases.resolve(input)
		return hilbish.runner.lua(cmdStr)
	end,
	validate = luaValidate
})

hilbish.runner.add('sh', {
	run = hilbish.runner.sh,
	validate = snail.validate
})
hilbish.runner.setCurrent 'hybrid'
