---
title: Completions
description: Tab completion for commands.
layout: doc
menu: 
  docs:
    parent: "Features"
---

Pressing Tab at the prompt will suggest completions for the current word.
Hilbish provides completions for commands that have a registered completion
handler. If no command-specific handler exists, it falls back to completing
file paths.

Completions can appear in two layouts depending on what the handler provides:

- *Grid*: items shown in a column grid, used for simple lists like filenames
  or subcommands
- *List*: items shown in a vertical list, used when completions have
  descriptions or aliases (common for command flags)

Some commands also provide a mixed view with both types at the same time.

The completion system is extensible: you can write handlers for any command and
Hilbish will use them when that command is on the line. See the
[hilbish.completions API](../../api/hilbish/hilbish.completions) for how to
register completions for your own commands.
