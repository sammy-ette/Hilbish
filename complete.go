package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/sammy-ette/hilbish/moonlight"
	"github.com/sammy-ette/hilbish/util"
)

var charEscapeMap = []string{
	"\"", "\\\"",
	"'", "\\'",
	"`", "\\`",
	" ", "\\ ",
	"(", "\\(",
	")", "\\)",
	"[", "\\[",
	"]", "\\]",
	"$", "\\$",
	"&", "\\&",
	"*", "\\*",
	">", "\\>",
	"<", "\\<",
	"|", "\\|",
}
var charEscapeMapInvert = invert(charEscapeMap)
var escapeReplaer = strings.NewReplacer(charEscapeMap...)
var escapeInvertReplaer = strings.NewReplacer(charEscapeMapInvert...)

func invert(m []string) []string {
	newM := make([]string, len(charEscapeMap))
	for i := range m {
		if (i+1)%2 == 0 {
			newM[i] = m[i-1]
			newM[i-1] = m[i]
		}
	}

	return newM
}

func splitForFile(str string) []string {
	split := []string{}
	sb := &strings.Builder{}
	quoted := false

	for i, r := range str {
		if r == '"' {
			quoted = !quoted
			sb.WriteRune(r)
		} else if r == ' ' && i > 0 && str[i-1] == '\\' {
			sb.WriteRune(r)
		} else if !quoted && r == ' ' {
			split = append(split, sb.String())
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
	}
	if strings.HasSuffix(str, " ") {
		split = append(split, "")
	}

	if sb.Len() > 0 {
		split = append(split, sb.String())
	}

	return split
}

func fileComplete(ctx string) ([]string, string) {
	q := splitForFile(ctx)
	path := ""
	if len(q) != 0 {
		path = q[len(q)-1]
	}

	return matchPath(path)
}

func dirComplete(query, ctx string) ([]string, string) {
	q := splitForFile(ctx)
	path := ""
	if len(q) != 0 {
		path = q[len(q)-1]
	}

	var completions []string

	fileCompletions, filePref := matchPath(path)
	for _, f := range fileCompletions {
		fullPath, _ := filepath.Abs(util.ExpandHome(path + strings.TrimPrefix(f, filePref)))
		fi, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			completions = append(completions, f)
		}
	}

	return completions, filePref
}

func binaryComplete(query, ctx string) ([]string, string) {
	q := splitForFile(ctx)
	query = ""
	if len(q) != 0 {
		query = q[len(q)-1]
	}

	var completions []string

	prefixes := []string{"./", "../", "/", "~/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(query, prefix) {
			fileCompletions, filePref := matchPath(query)
			if len(fileCompletions) != 0 {
				for _, f := range fileCompletions {
					fullPath, _ := filepath.Abs(util.ExpandHome(query + strings.TrimPrefix(f, filePref)))
					if err := util.FindExecutable(escapeInvertReplaer.Replace(fullPath), false, true); err != nil {
						continue
					}
					completions = append(completions, f)
				}
			}
			return completions, filePref
		}
	}

	// filter out executables, but in path
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		// search for an executable which matches our query string
		if matches, err := filepath.Glob(filepath.Join(dir, query+"*")); err == nil {
			// get basename from matches
			for _, match := range matches {
				// check if we have execute permissions for our match
				err := util.FindExecutable(match, true, false)
				if err != nil {
					continue
				}
				// get basename from match
				name := filepath.Base(match)
				// add basename to completions
				completions = append(completions, name)
			}
		}
	}

	// add lua registered commands to completions
	for cmdName := range cmds.Commands {
		if strings.HasPrefix(cmdName, query) {
			completions = append(completions, cmdName)
		}
	}

	completions = removeDupes(completions)

	return completions, query
}

func matchPath(query string) ([]string, string) {
	oldQuery := query
	query = strings.TrimPrefix(query, "\"")
	var entries []string
	var baseName string

	query = escapeInvertReplaer.Replace(query)
	path, _ := filepath.Abs(util.ExpandHome(filepath.Dir(query)))
	if string(query) == "" {
		// filepath base below would give us "."
		// which would cause a match of only dotfiles
		path, _ = filepath.Abs(".")
	} else if !strings.HasSuffix(query, string(os.PathSeparator)) {
		baseName = filepath.Base(query)
	}

	files, _ := os.ReadDir(path)
	for _, entry := range files {
		file, err := entry.Info()
		if err != nil {
			continue
		}

		if file.Mode()&os.ModeSymlink != 0 {
			// If the symlink is broken (or otherwise can't be resolved),
			// just keep treating it as the symlink entry itself
			if resolved, err := filepath.EvalSymlinks(filepath.Join(path, file.Name())); err == nil {
				if info, err := os.Lstat(resolved); err == nil {
					file = info
				}
			}
		}

		if strings.HasPrefix(file.Name(), baseName) {
			entry := file.Name()
			if file.IsDir() {
				entry = entry + string(os.PathSeparator)
			}
			if !strings.HasPrefix(oldQuery, "\"") {
				entry = escapeFilename(entry)
			}
			entries = append(entries, entry)
		}
	}
	return entries, baseName
}

func escapeFilename(fname string) string {
	return escapeReplaer.Replace(fname)
}

// @interface completions
// tab completions
// The completions interface provides functions to register and manage tab completions.
//
// ## Completer Function
//
// A function registered for a specific command scope with
// `hilbish.completions.add`. The scope string is `command.<name>` where `<name>` is
// the command being completed (e.g. `command.git`). The handler is called with three
// arguments:
//
// - `query` (string): The word the user is currently trying to complete. Use this to filter your items.
// - `ctx` (string): The full command line as a string.
// - `fields` (table): The command line split into fields by whitespace. `fields[1]` is the command name, `fields[2]` is the first argument, and so on.
//
// The handler must return two values: a table of *completion groups* and a prefix string.
// The prefix is usually just `query`.
//
// ## Completion Groups
//
// A completion group is a table with two fields: `type` and `items`.
// Multiple groups can be returned at once and Hilbish will display them together.
//
// *Grid*: items shown side by side in a grid. `items` is a list of strings:
//
// ```lua
// { type = 'grid', items = {'add', 'commit', 'push', 'pull'} }
// ```
//
// *List*: items shown in a vertical list with optional descriptions and aliases.
// Each entry in `items` can be a plain string or a table with these keys (all optional): `description`, `alias`, `display`.
//
// ```lua
//
//	{
//	  type = 'list',
//	  items = {
//	    ['--verbose'] = { description = 'enable verbose output', alias = '-v' },
//	    ['--output']  = { description = 'output file path' },
//	    '--dry-run',
//	  }
//	}
//
// ```
//
// ## Example
//
// Here is a full completer for a `sudo`-like command: it completes binaries when
// no argument has been typed yet, and falls back to file completion otherwise.
//
// ```lua
// hilbish.completions.add('command.sudo', function(query, ctx, fields)
//
//	if #fields == 0 then
//		-- complete for commands
//		local comps, pfx = hilbish.completions.bins(query, ctx, fields)
//		local compGroup = {
//			items = comps, -- our list of items to complete
//			type = 'grid' -- what our completions will look like.
//		}
//
//		return {compGroup}, pfx
//	end
//
//	-- otherwise just be boring and return files
//
//	local comps, pfx = hilbish.completions.files(query, ctx, fields)
//	local compGroup = {
//		items = comps,
//		type = 'grid'
//	}
//
//	return {compGroup}, pfx
//
// end)
// ```
func completionLoader(mlr *moonlight.Runtime) *moonlight.Table {
	exports := map[string]moonlight.Export{
		"bins":    {Function: hcmpBins, ArgNum: 3, Variadic: false},
		"call":    {Function: hcmpCall, ArgNum: 4, Variadic: false},
		"files":   {Function: hcmpFiles, ArgNum: 3, Variadic: false},
		"dirs":    {Function: hcmpDirs, ArgNum: 3, Variadic: false},
		"add":     {Function: hcmpAdd, ArgNum: 2, Variadic: false},
		"handler": {Function: hcmpHandler, ArgNum: 2, Variadic: false},
	}

	mod := moonlight.NewTable()
	mlr.SetExports(mod, exports)

	return mod
}

// @interface completions
// add(scope, cb)
// Registers a completion handler for the specified scope.
// A `scope` is expected to be `command.<cmd>`,
// replacing <cmd> with the name of the command (for example `command.git`).
// See the module introduction above for a full worked example, and the
// documentation for completions, under Features/Completions or `doc completions`,
// for more details.
// @param scope string
// @param cb fun(query:string,ctx:string,fields:table<string>):table,string
func hcmpAdd(mlr *moonlight.Runtime) error {
	scope, cb, err := util.HandleStrCallback(mlr)
	if err != nil {
		return err
	}
	luaCompletionsMu.Lock()
	luaCompletions[scope] = cb
	luaCompletionsMu.Unlock()

	return nil
}

// @interface completions
// bins
// Return binaries/executables based on the provided parameters.
// This function is meant to be used as a helper in a command completion handler,
// as shown in the module introduction above.
// @param query string Text the user is currently trying to complete.
// @param ctx string The full command line string.
// @param fields table The command line split into fields by whitespace.
// @return table<string> entries A list of entries.
// @return string prefix The prefix used for completions.
func hcmpBins(mlr *moonlight.Runtime) error {
	query, ctx, fds, err := getCompleteParams(mlr)
	if err != nil {
		return err
	}

	var _ []string = fds
	completions, pfx := binaryComplete(query, ctx)
	luaComps := moonlight.NewTable()

	for i, comp := range completions {
		luaComps.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(comp))
	}

	mlr.PushNext(moonlight.TableValue(luaComps), moonlight.StringValue(pfx))
	return nil
}

// @interface completions
// call(name, query, ctx, fields) -> completionGroups (table), prefix (string)
// Calls a completer function.
// This is mainly used to call a command completer, which will have a `name`
// in the form of `command.name`, example: `command.git`.
// @param name string The name of the completer to call, e.g. `command.git`.
// @param query string Text the user is currently trying to complete.
// @param ctx string The full command line string.
// @param fields table The command line split into fields by whitespace.
// @return table completionGroups A table of completion groups.
// @return string prefix
// @since 2.0.0
func hcmpCall(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(4); err != nil {
		return err
	}
	completer, err := mlr.StringArg(0)
	if err != nil {
		return err
	}
	query, err := mlr.StringArg(1)
	if err != nil {
		return err
	}
	ctx, err := mlr.StringArg(2)
	if err != nil {
		return err
	}
	fields, err := mlr.TableArg(3)
	if err != nil {
		return err
	}

	luaCompletionsMu.RLock()
	completecb, ok := luaCompletions[completer]
	luaCompletionsMu.RUnlock()
	if !ok {
		return errors.New("completer " + completer + " does not exist")
	}

	results, err := mlr.Call(moonlight.FunctionValue(completecb), moonlight.StringValue(query), moonlight.StringValue(ctx), moonlight.TableValue(fields))
	if err != nil {
		return err
	}

	mlr.PushNext(results...)
	return nil
}

// @interface completions
// files
// Returns file matches based on the provided parameters.
// This function is meant to be used as a helper in a command completion handler.
// @param query string Text the user is currently trying to complete.
// @param ctx string The full command line string.
// @param fields table The command line split into fields by whitespace.
// @return table<string> entries A list of entries.
// @return string prefix The prefix used for completions.
func hcmpFiles(mlr *moonlight.Runtime) error {
	query, ctx, fds, err := getCompleteParams(mlr)
	if err != nil {
		return err
	}

	var _ []string = fds
	var _ string = query
	completions, pfx := fileComplete(ctx)
	luaComps := moonlight.NewTable()

	for i, comp := range completions {
		luaComps.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(comp))
	}

	mlr.PushNext(moonlight.TableValue(luaComps), moonlight.StringValue(pfx))
	return nil
}

// @interface completions
// dirs(query, ctx, fields) -> entries (table), prefix (string)
// Returns directory matches based on the provided parameters.
// This function is meant to be used as a helper in a command completion handler.
// @param query string Text the user is currently trying to complete.
// @param ctx string The full command line string.
// @param fields table The command line split into fields by whitespace.
// @return table<string> entries A list of entries.
// @return string prefix The prefix used for completions.
// @since 3.0.0
func hcmpDirs(mlr *moonlight.Runtime) error {
	query, ctx, fds, err := getCompleteParams(mlr)
	if err != nil {
		return err
	}

	var _ []string = fds
	completions, pfx := dirComplete(query, ctx)
	luaComps := moonlight.NewTable()

	for i, comp := range completions {
		luaComps.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(comp))
	}

	mlr.PushNext(moonlight.TableValue(luaComps), moonlight.StringValue(pfx))
	return nil
}

// @interface completions
// handler(line, pos)
// This function contains the general completion handler for Hilbish.
// This function handles completion of everything,
// which includes calling other command handlers, binaries, and files.
// This function can be overridden to supply a custom handler. Note that alias resolution is required to be done in this function.
// @param line string The current Hilbish command line
// @param pos number Numerical position of the cursor
// @return string prefix The common prefix of all completion items
// @return table completionGroups A list of completion groups
// @since 2.0.0
/*
@example
-- stripped down version of the default implementation
function hilbish.completions.handler(line, pos)
	local query = fields[#fields]

	if #fields == 1 then
		-- call bins handler here
	else
		-- call command completer or files completer here
	end
end
@example
*/
func hcmpHandler(mlr *moonlight.Runtime) error {
	return nil
}

func getCompleteParams(mlr *moonlight.Runtime) (string, string, []string, error) {
	if err := mlr.CheckNArgs(3); err != nil {
		return "", "", []string{}, err
	}
	query, err := mlr.StringArg(0)
	if err != nil {
		return "", "", []string{}, err
	}
	ctx, err := mlr.StringArg(1)
	if err != nil {
		return "", "", []string{}, err
	}
	fields, err := mlr.TableArg(2)
	if err != nil {
		return "", "", []string{}, err
	}

	var fds []string
	moonlight.ForEach(fields, func(k moonlight.Value, v moonlight.Value) {
		if v.Type() == moonlight.StringType {
			fds = append(fds, v.AsString())
		}
	})

	return query, ctx, fds, err
}
