// library for custom commands
/*
Commander is the library which handles Hilbish commands. This makes
the user able to add Lua-written commands to their shell without making
a separate script in a bin folder. Instead, you may simply use the Commander
library in your Hilbish config.

```lua
local commander = require 'commander'

commander.register('hello', function(args, sinks)
	sinks.out:writeln 'Hello world!'
end)
```

In this example, a command with the name of `hello` is created
that will print `Hello world!` to output. One question you may
have is: What is the `sinks` parameter?<nl>
The `sinks` parameter is a table with 3 keys: `input`, `out`, and `err`.
There is an `in` alias to `input`, but it requires using the string accessor syntax (`sinks['in']`)
as `in` is also a Lua keyword, so `input` is preferred for use.
All of them are a @Sink.
In the future, `sinks.in` will be removed.<nl>
- `in` is the standard input. You may use the read functions on this sink to get input from the user.
- `out` is standard output. This is usually where command output should go.
- `err` is standard error. This sink is for writing errors, as the name would suggest.
*/
package commander

import (
	"hilbish/golibs/bait"
	"hilbish/util"

	"github.com/arnodel/golua/lib/packagelib"
	rt "github.com/arnodel/golua/runtime"
)

type Commander struct {
	Events   *bait.Bait
	Loader   packagelib.Loader
	Commands map[string]*rt.Closure
}

func New(rtm *rt.Runtime) *Commander {
	c := &Commander{
		Events:   bait.New(rtm),
		Commands: make(map[string]*rt.Closure),
	}
	c.Loader = packagelib.Loader{
		Load: c.loaderFunc,
		Name: "commander",
	}

	return c
}

func (c *Commander) loaderFunc(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]util.LuaExport{
		"register":   util.LuaExport{c.cregister, 2, false},
		"deregister": util.LuaExport{c.cderegister, 1, false},
		"registry":   util.LuaExport{c.cregistry, 0, false},
	}
	mod := rt.NewTable()
	util.SetExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}

// register(name, cb)
// Adds a new command with the given `name`. When Hilbish has to run a command with a name,
// it will run the function providing the arguments and sinks.
// #param name string Name of the command
// #param cb function Callback to handle command invocation
/*
#example
-- When you run the command `hello` in the shell, it will print `Hello world`.
-- If you run it with, for example, `hello Hilbish`, it will print 'Hello Hilbish'
commander.register('hello', function(args, sinks)
	local name = 'world'
	if #args > 0 then name = args[1] end

	sinks.out:writeln('Hello ' .. name)
end)
#example
*/
func (c *Commander) cregister(t *rt.Thread, ct *rt.GoCont) (rt.Cont, error) {
	cmdName, cmd, err := util.HandleStrCallback(t, ct)
	if err != nil {
		return nil, err
	}

	c.Commands[cmdName] = cmd

	return ct.Next(), err
}

// deregister(name)
// Removes the named command. Note that this will only remove Commander-registered commands.
// #param name string Name of the command to remove.
func (c *Commander) cderegister(t *rt.Thread, ct *rt.GoCont) (rt.Cont, error) {
	if err := ct.Check1Arg(); err != nil {
		return nil, err
	}
	cmdName, err := ct.StringArg(0)
	if err != nil {
		return nil, err
	}

	delete(c.Commands, cmdName)

	return ct.Next(), err
}

// registry() -> table
// Returns all registered commanders. Returns a list of tables with the following keys:
// - `exec`: The function used to run the commander. Commanders require args and sinks to be passed.
// #returns table
func (c *Commander) cregistry(t *rt.Thread, ct *rt.GoCont) (rt.Cont, error) {
	registryLua := rt.NewTable()
	for cmdName, cmd := range c.Commands {
		cmdTbl := rt.NewTable()
		cmdTbl.Set(rt.StringValue("exec"), rt.FunctionValue(cmd))

		registryLua.Set(rt.StringValue(cmdName), rt.TableValue(cmdTbl))
	}

	return ct.PushingNext1(t.Runtime, rt.TableValue(registryLua)), nil
}
