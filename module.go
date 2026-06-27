package main

import (
	"fmt"
	"hilbish/moonlight"
	"plugin"
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
with a signature like so: `func(*moonlight.Runtime) moonlight.Value`.

Here is some code for an example plugin:
```go
package main

import (
	"github.com/sammy-ette/hilbish/moonlight"
)

func Loader(rtm *moonlight.Runtime) moonlight.Value {
	return moonlight.StringValue("hello world!")
}
```

This can be compiled with `go build -buildmode=plugin plugin.go`.
If you attempt to require and print the result (`print(require 'plugin')`), it will show "hello world!"
*/
func moduleLoader(mlr *moonlight.Runtime) *moonlight.Table {
	exports := map[string]moonlight.Export{
		"load": {Function: moduleLoad, ArgNum: 2, Variadic: false},
	}

	mod := moonlight.NewTable()
	mlr.SetExports(mod, exports)

	return mod
}

// #interface module
// load(path)
// Loads a module at the designated `path`.
// It will throw if any error occurs.
// #param path string
func moduleLoad(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	value, err := p.Lookup("Loader")
	if err != nil {
		return err
	}

	loader, ok := value.(func(*moonlight.Runtime) moonlight.Value)
	if !ok {
		return fmt.Errorf("module has wrong function signature: should be func(*moonlight.Runtime) moonlight.Value")
	}

	val := loader(mlr)
	mlr.PushNext(val)

	return nil
}
