package main

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
)

var aliases *aliasModule

type aliasModule struct {
	aliases map[string]string
	mu      *sync.RWMutex
}

// initialize aliases map
func newAliases() *aliasModule {
	return &aliasModule{
		aliases: make(map[string]string),
		mu:      &sync.RWMutex{},
	}
}

func (a *aliasModule) Add(alias, cmd string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.aliases[alias] = cmd
}

func (a *aliasModule) All() map[string]string {
	return a.aliases
}

func (a *aliasModule) Delete(alias string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.aliases, alias)
}

func (a *aliasModule) Resolve(cmdstr string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	arg, _ := regexp.Compile(`[\\]?%\d+`)

	args, _ := splitInput(cmdstr)
	if len(args) == 0 {
		// this shouldnt reach but...????
		return cmdstr
	}

	visited := make(map[string]bool)
	for a.aliases[args[0]] != "" {
		if visited[args[0]] {
			// cyclic alias definition (eg foo -> bar -> foo), stop resolving
			break
		}
		visited[args[0]] = true

		alias := a.aliases[args[0]]
		alias = arg.ReplaceAllStringFunc(alias, func(a string) string {
			idx, _ := strconv.Atoi(a[1:])
			if strings.HasPrefix(a, "\\") || idx == 0 {
				return strings.TrimPrefix(a, "\\")
			}

			if idx+1 > len(args) {
				return a
			}
			val := args[idx]
			args = cut(args, idx)
			cmdstr = strings.Join(args, " ")

			return val
		})

		cmdstr = alias + strings.TrimPrefix(cmdstr, args[0])
		cmdArgs, _ := splitInput(cmdstr)
		args = cmdArgs
	}

	return cmdstr
}

// lua section

// #interface aliases
// command aliasing
// The alias interface deals with all command aliases in Hilbish.
func (a *aliasModule) Loader(rtm *rt.Runtime) *rt.Table {
	// create a lua module with our functions
	hshaliasesLua := map[string]util.LuaExport{
		"add":     {Function: hlalias, ArgNum: 2, Variadic: false},
		"list":    {Function: a.luaList, ArgNum: 0, Variadic: false},
		"del":     {Function: a.luaDelete, ArgNum: 1, Variadic: false},
		"resolve": {Function: a.luaResolve, ArgNum: 1, Variadic: false},
	}

	mod := rt.NewTable()
	util.SetExports(rtm, mod, hshaliasesLua)

	return mod
}

// #interface aliases
// add(alias, cmd)
// This is an alias (ha) for the [hilbish.alias](../#alias) function.
// --- @param alias string
// --- @param cmd string
func _hlalias() {}

// #interface aliases
// list() -> table[string, string]
// Get a table of all aliases, with string keys as the alias and the value as the command.
// #returns table[string, string]
/*
#example
hilbish.aliases.add('hi', 'echo hi')

local aliases = hilbish.aliases.list()
-- -> {hi = 'echo hi'}
#example
*/
func (a *aliasModule) luaList(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	aliasesList := rt.NewTable()
	for k, v := range a.All() {
		aliasesList.Set(rt.StringValue(k), rt.StringValue(v))
	}

	return c.PushingNext1(t.Runtime, rt.TableValue(aliasesList)), nil
}

// #interface aliases
// delete(name)
// Removes an alias.
// #param name string
func (a *aliasModule) luaDelete(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	alias, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	a.Delete(alias)

	return c.Next(), nil
}

// #interface aliases
// resolve(alias) -> string?
// Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.
// #param alias string
// #returns string
func (a *aliasModule) luaResolve(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	alias, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	resolved := a.Resolve(alias)

	return c.PushingNext1(t.Runtime, rt.StringValue(resolved)), nil
}
