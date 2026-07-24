-- @module hilbish.runner
--- interactive command executor
--- The runner interface contains functions that allow the user to change
--- how Hilbish interprets interactive input.
--- Users can add and change the default runner for interactive input to any
--- language or script of their choosing. A good example is using it to
--- write commands in Fennel.
---
--- A runner is a table with `run` and `validate` functions; see the
--- `Runner` type below. `run` returns a `RunnerResult`, also detailed below.
---
--- Here is a simple example of a fennel runner. It falls back to
--- shell script if fennel eval has an error.
---
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
local fs = require 'fs'
local snail = require 'snail'
local currentRunner = 'hybrid'
local runners = {}

---@diagnostic disable-next-line: missing-fields
hilbish.runner = {}

--- A table describing how to run and validate interactive input for a runner mode.
--- @class Runner
--- @field run fun(input: string): RunnerResult Evaluates the input and returns a result table.
--- @field validate fun(input: string): boolean Checks whether the input is complete and ready to run. Return `false` to prompt the user for more input (continuation), or `true` to proceed.

--- The table returned by a `Runner`'s `run` function. All fields are optional;
--- only set the ones relevant to the runner (if there isn't an error, just omit `err`).
--- @class RunnerResult
--- @field exitCode? number Exit code of the command.
--- @field input? string The text input of the user. Used by Hilbish to append extra input if more is requested.
--- @field err? string A string that represents an error from the runner. This should only be set on syntax errors. It can be set to a few special values for Hilbish to throw the right hooks and show a better looking message: `<command>: not-found` will throw a `command.not-found` hook, `<command>: not-executable` will throw a `command.not-executable` hook.
--- @field continue? boolean Whether Hilbish should prompt the user for more input.
--- @field newline? boolean Whether a newline should be added at the end of `input`.

--- Get a runner by name. Throws an error if the runner does not exist.
--- @param name string Name of the runner to retrieve.
--- @return Runner runner
--- @since 2.0.0
function hilbish.runner.get(name)
	local r = runners[name]

	if not r then
		error(string.format('runner %s does not exist', name))
	end

	return r
end

--- Adds a runner to the table of available runners. Errors if a runner
--- with `name` already exists. Use `set` to overwrite an existing runner.
--- @param name string Unique name for the runner.
--- @param runner Runner
--- @see set
--- @since 3.0.0
function hilbish.runner.add(name, runner)
	if runners[name] then
		error(string.format('runner %s already exists', name))
	end

	hilbish.runner.set(name, runner)
end

--- Sets (or replaces) a runner by name, without checking if one already exists.
--- @param name string Name of the runner to set.
--- @param runner Runner
--- @since 3.0.0
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

--- Runs `cmd` using the named runner, or the current runner if `runnerName` is not given.
--- @param cmd string The command string to run.
--- @param runnerName? string The name of the runner to use. Defaults to the current runner.
--- @return RunnerResult result
--- @since 2.0.0
function hilbish.runner.exec(cmd, runnerName)
	if not runnerName then runnerName = currentRunner end

	local r = hilbish.runner.get(runnerName)

	return r.run(cmd)
end

--- Sets Hilbish's active runner mode by name. Errors if the runner does not exist.
--- @param name string Name of the runner to make active.
--- @since 2.0.0
function hilbish.runner.setCurrent(name)
	hilbish.runner.get(name) -- throws if it doesnt exist.
	currentRunner = name
end

--- Returns the name of the currently active runner.
--- @return string name The name of the current runner.
--- @since 2.1.0
function hilbish.runner.getCurrent()
	return currentRunner
end

local function finishExec(exitCode, input, priv)
	hilbish.exitCode = exitCode
	bait.throw('command.exit', exitCode, input, priv) -- see nature/hooks.lua
end

--- Prompts for continued input from the user.
--- @param prev string
--- @param newline boolean
--- @return string|nil input
--- @private
--- @since 3.0.0
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
--- @default priv? false
--- @since 3.0.0
function hilbish.runner.run(input, priv)
	bait.throw('command.preprocess', input) -- see nature/hooks.lua
	local processed = hilbish.processors.execute(input, {
		skip = hilbish.opts.processorSkipList
	})

	if processed.modifiers.private or processed.modifiers.priv then
		priv = true
	end

	if not processed.continue then
		finishExec(0, '', true)
		return
	end

	local runner = hilbish.runner.get(processed.modifiers.runner or currentRunner)
	local oldDir = hilbish.cwd()
	
	::rerun::
	local command
	if processed.modifiers.alias == false then
		command = processed.command
	else
		command = hilbish.aliases.resolve(processed.command)
	end

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

	bait.throw('command.preexec', processed.command, command) -- see nature/hooks.lua

	local function cd(dir)
		return pcall(fs.cd, dir)
	end

	local function cdToOld()
		if processed.modifiers.dir then
			local ok, err = cd(oldDir)
			if not ok then
				io.stderr:write('hilbish: failed to restore directory: ' .. tostring(err) .. '\n')
			end
		end
	end

	if processed.modifiers.dir then
		local ok, err = cd(processed.modifiers.dir)
		if not ok then
			io.stderr:write('hilbish: @dir: ' .. tostring(err) .. '\n')
			finishExec(1, '', true)
			return
		end
	end

	local ok, out = pcall(runner.run, processed.command)
	if not ok then
		io.stderr:write(out .. '\n')
		finishExec(124, out.input, priv)
		cdToOld()
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
	cdToOld()
end

--- Runs `input` as a shell script using Hilbish's built-in shell interpreter.
--- @param input string The shell script to run.
--- @return RunnerResult result
--- @since 2.0.0
function hilbish.runner.sh(input)
	return hilbish.snail:run(input)
end

--- Evaluates `input` as Lua code. Equivalent to `load(input)()`, shaped
--- for the runner interface.
--- @param input string The Lua code to evaluate.
--- @return RunnerResult result
--- @since 2.0.0
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

hilbish.runner.add('hybrid', {
	run = function(input)
		local cmdStr = hilbish.aliases.resolve(input)

		local res = hilbish.runner.lua(cmdStr)
		if not res.err then
			return res
		end

		return hilbish.runner.sh(input)
	end,
	validate = function(input)
		return luaValidate(input) or snail.validate(input)
	end
})

hilbish.runner.add('hybridRev', {
	run = function(input)
		local res = hilbish.runner.sh(input)
		if not res.err then
			return res
		end

		local cmdStr = hilbish.aliases.resolve(input)
		return hilbish.runner.lua(cmdStr)
	end,
	validate = function(input)
		return snail.validate(input) or luaValidate(input)
	end
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
