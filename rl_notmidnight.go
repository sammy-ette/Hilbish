//go:build !midnight

package main

import (
	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
	"github.com/maxlandon/readline"
)

func setupTabCompleter(rl *readline.Readline) {
	rl.TabCompleter = func(line []rune, pos int, _ readline.DelayedTabContext) (string, []*readline.CompletionGroup) {
		term := rt.NewTerminationWith(l.UnderlyingRuntime().MainThread().CurrentCont(), 2, false)
		compHandle := hshMod.Get(rt.StringValue("completion")).AsTable().Get(rt.StringValue("handler"))
		err := rt.Call(l.UnderlyingRuntime().MainThread(), compHandle, []rt.Value{rt.StringValue(string(line)),
			rt.IntValue(int64(pos))}, term)

		var compGroups []*readline.CompletionGroup
		if err != nil {
			return "", compGroups
		}

		luaCompGroups := term.Get(0)
		luaPrefix := term.Get(1)

		if luaCompGroups.Type() != rt.TableType {
			return "", compGroups
		}

		groups := luaCompGroups.AsTable()
		// prefix is optional
		pfx, _ := luaPrefix.TryString()

		util.ForEach(groups, func(key rt.Value, val rt.Value) {
			if key.Type() != rt.IntType || val.Type() != rt.TableType {
				return
			}

			valTbl := val.AsTable()
			luaCompType := valTbl.Get(rt.StringValue("type"))
			luaCompItems := valTbl.Get(rt.StringValue("items"))

			if luaCompType.Type() != rt.StringType || luaCompItems.Type() != rt.TableType {
				return
			}

			items := []string{}
			itemDescriptions := make(map[string]string)
			itemDisplays := make(map[string]string)
			itemAliases := make(map[string]string)

			util.ForEach(luaCompItems.AsTable(), func(lkey rt.Value, lval rt.Value) {
				if keytyp := lkey.Type(); keytyp == rt.StringType {
					// TODO: remove in 3.0
					// ['--flag'] = {'description', '--flag-alias'}
					// OR
					// ['--flag'] = {description = '', alias = '', display = ''}
					itemName, ok := lkey.TryString()
					vlTbl, okk := lval.TryTable()
					if !ok && !okk {
						// TODO: error
						return
					}

					items = append(items, itemName)
					itemDescription, ok := vlTbl.Get(rt.IntValue(1)).TryString()
					if !ok {
						// if we can't get it by number index, try by string key
						itemDescription, _ = vlTbl.Get(rt.StringValue("description")).TryString()
					}
					itemDescriptions[itemName] = itemDescription

					// display
					if itemDisplay, ok := vlTbl.Get(rt.StringValue("display")).TryString(); ok {
						itemDisplays[itemName] = itemDisplay
					}

					itemAlias, ok := vlTbl.Get(rt.IntValue(2)).TryString()
					if !ok {
						// if we can't get it by number index, try by string key
						itemAlias, _ = vlTbl.Get(rt.StringValue("alias")).TryString()
					}
					itemAliases[itemName] = itemAlias
				} else if keytyp == rt.IntType {
					vlStr, ok := lval.TryString()
					if !ok {
						// TODO: error
						return
					}
					items = append(items, vlStr)
				} else {
					// TODO: error
					return
				}
			})

			var dispType readline.TabDisplayType
			switch luaCompType.AsString() {
			case "grid":
				dispType = readline.TabDisplayGrid
			case "list":
				dispType = readline.TabDisplayList
				// need special cases, will implement later
				//case "map": dispType = readline.TabDisplayMap
			}

			compGroups = append(compGroups, &readline.CompletionGroup{
				DisplayType:  dispType,
				Aliases:      itemAliases,
				Descriptions: itemDescriptions,
				ItemDisplays: itemDisplays,
				Suggestions:  items,
				TrimSlash:    false,
				NoSpace:      true,
			})
		})

		return pfx, compGroups
	}
}
