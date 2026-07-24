---
title: Aliases
description: Short names for longer commands.
layout: doc
weight: -20
menu: 
  docs:
    parent: "Features"
---

Aliases are short names that expand to longer commands when you run them.
You type the short name, Hilbish runs the full
command, and the short name is what gets saved to history.

Aliases also support argument substitution. If an alias uses `%1`, `%2`, etc.,
those get replaced with the corresponding argument you pass. For example, an
alias using `%1` that you call as `myalias foo` will substitute `foo` in for
`%1` when running the full command.

To skip alias expansion for a single command, use the
[`@alias=false` modifier](../modifiers).

Aliases are defined in your config file. See the
[hilbish.aliases API](../../api/hilbish/hilbish.aliases) for how to add,
list, and remove them.
