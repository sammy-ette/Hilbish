package main

import (
	"plugin"

	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
)

// #interface module
// native module loading
// #field paths A list of paths to search when loading native modules. This is in the style of Lua search paths and will be used when requiring native modules. Example: `?.so;?/?.so`
/*
The hilbish.module interface provides a function to load
Hilbish plugins/modules. Hilbish modules are Go-written
plugins (see https://pkg.go.dev/plugin) that are used to add functionality
to Hilbish that cannot be written in Lua for any reason.

Note that you don't ever need to use the load function that is here as
modules can be loaded with a `require` call like Lua C modules, and the
search paths can be changed with the `paths` property here.

To make a valid native module, the Go plugin has to export a Loader function
with a signature like so: `func(*rt.Runtime) rt.Value`.

`rt` in this case refers to the Runtime type at
https://pkg.go.dev/github.com/arnodel/golua@master/runtime#Runtime

Hilbish uses this package as its Lua runtime. You will need to read
it to use it for a native plugin.

Here is some code for an example plugin:
```go
package main

import (
	rt "github.com/arnodel/golua/runtime"
)

func Loader(rtm *rt.Runtime) rt.Value {
	return rt.StringValue("hello world!")
}
```

This can be compiled with `go build -buildmode=plugin plugin.go`.
If you attempt to require and print the result (`print(require 'plugin')`), it will show "hello world!"
*/
func moduleLoader(rtm *rt.Runtime) *rt.Table {
	exports := map[string]util.LuaExport{
		"load": {moduleLoad, 2, false},
	}

	mod := rt.NewTable()
	util.SetExports(rtm, mod, exports)

	return mod
}

// #interface module
// load(path)
// Loads a module at the designated `path`.
// It will throw if any error occurs.
// #param path string 
func moduleLoad(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(1); err != nil {
		return nil, err
	}
	
	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	value, err := p.Lookup("Loader")
	if err != nil {
		return nil, err
	}

	loader, ok := value.(func(*rt.Runtime) rt.Value)
	if !ok {
		return nil, nil
	}

	val := loader(t.Runtime)

	return c.PushingNext1(t.Runtime, val), nil
}
