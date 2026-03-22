-- @module hilbish
local bait = require 'bait'
local snail = require 'snail'

hilbish.snail = snail.new()
hilbish.snail:run 'true' -- to "initialize" snail
bait.catch('hilbish.cd', function(path)
	hilbish.snail:dir(path)
end)
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
