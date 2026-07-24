---
title: Runner Mode
description: Customize how Hilbish interprets interactive input.
layout: doc
menu: 
  docs:
    parent: "Features"
---

The runner is what Hilbish uses to interpret the text you type at the prompt.
By default it uses a hybrid mode: it first tries to evaluate your input as Lua,
and if that fails, it runs it as a shell command. This means you can freely mix
shell commands and Lua expressions at the prompt.

The built-in modes are:

- **hybrid** — try Lua first, then shell (the default)
- **hybridRev** — try shell first, then Lua
- **sh** — shell only
- **lua** — Lua only (turns the prompt into a Lua REPL)

Switching to a different mode changes how *every* line you type is interpreted.
If you want to run a single command with a different runner without changing the
global mode, use the [`@runner` modifier](../modifiers) instead.

You can also write entirely custom runners — for example, to use a different
language like Fennel at the prompt. See the
[hilbish.runner API](../../api/hilbish/hilbish.runner) for how runners work and
how to register your own.
