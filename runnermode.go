package main

import (
	"hilbish/moonlight"
)

// #interface runner
// interactive command runner customization
/* The runner interface contains functions that allow the user to change
how Hilbish interprets interactive input.
Users can add and change the default runner for interactive input to any
language or script of their choosing. A good example is using it to
write command in Fennel.

Runners are functions that evaluate user input. The default runners in
Hilbish can run shell script and Lua code.

A runner is passed the input and has to return a table with these values.
All are not required, only the useful ones the runner needs to return.
(So if there isn't an error, just omit `err`.)

- `exitCode` (number): Exit code of the command
- `input` (string): The text input of the user. This is used by Hilbish to append extra input, in case
more is requested.
- `err` (string): A string that represents an error from the runner.
This should only be set when, for example, there is a syntax error.
It can be set to a few special values for Hilbish to throw the right
hooks and have a better looking message.
	- `<command>: not-found` will throw a `command.not-found` hook
	based on what `<command>` is.
	- `<command>: not-executable` will throw a `command.not-executable` hook.
- `continue` (boolean): Whether Hilbish should prompt the user for no input
- `newline` (boolean): Whether a newline should be added at the end of `input`.

Here is a simple example of a fennel runner. It falls back to
shell script if fennel eval has an error.
```lua
local fennel = require 'fennel'

hilbish.runnerMode(function(input)
	local ok = pcall(fennel.eval, input)
	if ok then
		return {
			input = input
		}
	end

	return hilbish.runner.sh(input)
end)
```
*/
func runnerModeLoader(mlr *moonlight.Runtime) *moonlight.Table {
	exports := map[string]moonlight.Export{
		"lua": {luaRunner, 1, false},
	}

	mod := moonlight.NewTable()
	mlr.SetExports(mod, exports)

	return mod
}

// #interface runner
// lua(cmd)
// Evaluates `cmd` as Lua input. This is the same as using `dofile`
// or `load`, but is appropriated for the runner interface.
// #param cmd string
func luaRunner(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	cmd, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	input, exitCode, err := handleLua(cmd)
	var luaErr moonlight.Value = moonlight.NilValue
	if err != nil {
		luaErr = moonlight.StringValue(err.Error())
	}
	runnerRet := moonlight.NewTable()
	runnerRet.SetField("input", moonlight.StringValue(input))
	runnerRet.SetField("exitCode", moonlight.IntValue(int64(exitCode)))
	runnerRet.SetField("err", luaErr)

	mlr.PushNext1(moonlight.TableValue(runnerRet))
	return nil
}
