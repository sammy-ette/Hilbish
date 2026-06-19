// filesystem interaction and functionality library
/*
The fs module provides filesystem functions to Hilbish. While Lua's standard
library has some I/O functions, they're missing a lot of the basics. The `fs`
library offers more functions and will work on any operating system Hilbish does.
#field pathSep The operating system's path separator.
*/
package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"hilbish/moonlight"
	"hilbish/util"
)

func Loader(mlr *moonlight.Runtime) moonlight.Value {
	exports := map[string]moonlight.Export{
		"cd":         {Function: fcd, ArgNum: 1, Variadic: false},
		"executable": {Function: fexecutable, ArgNum: 1, Variadic: false},
		"mkdir":      {Function: fmkdir, ArgNum: 2, Variadic: false},
		"stat":       {Function: fstat, ArgNum: 1, Variadic: false},
		"readdir":    {Function: freaddir, ArgNum: 1, Variadic: false},
		"abs":        {Function: fabs, ArgNum: 1, Variadic: false},
		"basename":   {Function: fbasename, ArgNum: 1, Variadic: false},
		"dir":        {Function: fdir, ArgNum: 1, Variadic: false},
		"glob":       {Function: fglob, ArgNum: 1, Variadic: false},
		"join":       {Function: fjoin, ArgNum: 0, Variadic: true},
		// "pipe":       {Function: fpipe, ArgNum: 0, Variadic: false},
	}
	mod := moonlight.NewTable()
	mlr.SetExports(mod, exports)
	mod.Set(moonlight.StringValue("pathSep"), moonlight.StringValue(string(os.PathSeparator)))
	mod.Set(moonlight.StringValue("pathListSep"), moonlight.StringValue(string(os.PathListSeparator)))

	return moonlight.TableValue(mod)
}

// abs(path) -> string
// Returns an absolute version of the `path`.
// This can be used to resolve short paths like `..` to `/home/user`.
// #param path string
// #returns string
func fabs(mlr *moonlight.Runtime) error {
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	path = util.ExpandHome(path)

	abspath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	mlr.PushNext1(moonlight.StringValue(abspath))
	return nil
}

// basename(path) -> string
// Returns the "basename," or the last part of the provided `path`. If path is empty,
// `.` will be returned.
// #param path string Path to get the base name of.
// #returns string
func fbasename(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	mlr.PushNext1(moonlight.StringValue(filepath.Base(path)))
	return nil
}

// cd(dir)
// Changes Hilbish's directory to `dir`.
// #param dir string Path to change directory to.
func fcd(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	path = util.ExpandHome(strings.TrimSpace(path))
	oldWd, _ := os.Getwd()

	abspath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	err = os.Chdir(path)
	if err != nil {
		return err
	}

	baitMod := mlr.MustDoString("return require 'bait'").AsTable()
	throw := baitMod.Get(moonlight.StringValue("throw"))
	mlr.Call1(throw, moonlight.StringValue("hilbish.cd"), moonlight.StringValue(abspath), moonlight.StringValue(oldWd))

	return nil
}

// dir(path) -> string
// Returns the directory part of `path`. If a file path like
// `~/Documents/doc.txt` then this function will return `~/Documents`.
// #param path string Path to get the directory for.
// #returns string
func fdir(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	mlr.PushNext1(moonlight.StringValue(filepath.Dir(path)))
	return nil
}

// executable(path) -> boolean
// Checks if `path` is an executable file.
// #param path string
// #returns boolean
func fexecutable(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	err = util.FindExecutable(path, true, false)
	if err != nil {
		mlr.PushNext1(moonlight.BoolValue(false))
	} else {
		mlr.PushNext1(moonlight.BoolValue(true))
	}

	return nil
}

// glob(pattern) -> matches (table)
// Match all files based on the provided `pattern`.
// For the syntax' refer to Go's filepath.Match function: https://pkg.go.dev/path/filepath#Match
// #param pattern string Pattern to compare files with.
// #returns table A list of file names/paths that match.
/*
#example
--[[
	Within a folder that contains the following files:
	a.txt
	init.lua
	code.lua
	doc.pdf
]]--
local matches = fs.glob './*.lua'
print(matches)
-- -> {'init.lua', 'code.lua'}
#example
*/
func fglob(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	pattern, err := mlr.StringArg(0)
	if err != nil {
		return err
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	luaMatches := moonlight.NewTable()

	for i, match := range matches {
		luaMatches.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(match))
	}

	mlr.PushNext1(moonlight.TableValue(luaMatches))
	return nil
}

// join(...path) -> string
// Takes any list of paths and joins them based on the operating system's path separator.
// #param path ...string Paths to join together
// #returns string The joined path.
/*
#example
-- This prints the directory for Hilbish's config!
print(fs.join(hilbish.userDir.config, 'hilbish'))
-- -> '/home/user/.config/hilbish' on Linux
#example
*/
func fjoin(mlr *moonlight.Runtime) error {
	strs := make([]string, len(mlr.Etc()))
	for i, v := range mlr.Etc() {
		if v.Type() != moonlight.StringType {
			// +2; go indexes of 0 and first arg from above
			return fmt.Errorf("bad argument #%d to join (expected string, got %s)", i+1, v.TypeName())
		}
		strs[i] = v.AsString()
	}

	res := filepath.Join(strs...)

	mlr.PushNext1(moonlight.StringValue(res))
	return nil
}

// mkdir(name, recursive)
// Creates a new directory with the provided `name`.
// With `recursive`, mkdir will create parent directories.
// #param name string Name of the directory
// #param recursive boolean Whether to create parent directories for the provided name
/*
#example
-- This will create the directory foo, then create the directory bar in the
-- foo directory. If recursive is false in this case, it will fail.
fs.mkdir('./foo/bar', true)
#example
*/
func fmkdir(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	recursive, err := mlr.BoolArg(1)
	if err != nil {
		return err
	}
	path = util.ExpandHome(strings.TrimSpace(path))

	if recursive {
		err = os.MkdirAll(path, 0744)
	} else {
		err = os.Mkdir(path, 0744)
	}
	if err != nil {
		return err
	}

	return nil
}

// pipe() -> file*, file*
// Returns a pair of connected files, also known as a pipe.
// The type returned is a Lua file, same as returned from `io` functions, like `io.open`.
// #returns file*
// #returns file*
// func fpipe(mlr *moonlight.Runtime) error {
// 	rf, wf, err := os.Pipe()
// 	if err != nil {
// 		return err
// 	}

// 	rfLua := iolib.NewFile(rf, 0)
// 	wfLua := iolib.NewFile(wf, 0)

// 	mlr.PushNext(rfLua.Value(mlr), wfLua.Value(mlr))

// 	return nil
// }

// readdir(path) -> table[string]
// Returns a list of all files and directories in the provided path.
// #param dir string
// #returns table
func freaddir(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	dir, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	dir = util.ExpandHome(dir)
	names := moonlight.NewTable()

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for i, entry := range dirEntries {
		names.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(entry.Name()))
	}

	mlr.PushNext1(moonlight.TableValue(names))
	return nil
}

// stat(path) -> {}
// Returns the information about a given `path`.
// The returned table contains the following values:
// name (string) - Name of the path
// size (number) - Size of the path in bytes
// mode (string) - Unix permission mode in an octal format string (with leading 0)
// isDir (boolean) - If the path is a directory
// #param path string
// #returns table
/*
#example
local inspect = require 'inspect'

local stat = fs.stat '~'
print(inspect(stat))
--[[
Would print the following:
{
  isDir = true,
  mode = "0755",
  name = "username",
  size = 12288
}
]]--
#example
*/
func fstat(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}
	path, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	path = util.ExpandHome(path)

	pathinfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	statTbl := moonlight.NewTable()
	statTbl.Set(moonlight.StringValue("name"), moonlight.StringValue(pathinfo.Name()))
	statTbl.Set(moonlight.StringValue("size"), moonlight.IntValue(pathinfo.Size()))
	statTbl.Set(moonlight.StringValue("mode"), moonlight.StringValue("0"+strconv.FormatInt(int64(pathinfo.Mode().Perm()), 8)))
	statTbl.Set(moonlight.StringValue("isDir"), moonlight.BoolValue(pathinfo.IsDir()))

	mlr.PushNext1(moonlight.TableValue(statTbl))
	return nil
}
