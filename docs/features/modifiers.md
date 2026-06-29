---
title: Modifiers
description: Inline flags for one-off command behavior.
layout: doc
menu: 
  docs:
    parent: "Features"
---

Modifiers are flags placed at the start of a line of input which change how
that one command is run, without affecting any other command. They look like
`@name` or `@name=value`, and one or more of them can be put before a command,
separated by spaces.

```
@priv echo this won't be saved to history
@dir=~/Downloads ls
@runner=lua print('hi')
```

A modifier with no `=value` is treated as the boolean `true`. A value of
`true` or `false` is converted to an actual boolean, and anything else is
kept as a string.

``` =html
<hr class="my-4">
```

### `priv` / `private`

#### Value: `boolean`

Runs the command without saving it to history, the same as if
`hilbish.opts.history` were temporarily disabled for just that command.

```
@priv curl -H "Authorization: Bearer secret" https://example.com
```

``` =html
<hr class="my-4">
```

### `dir`

Changes the shell's directory to `dir` only for the duration of the command,
then restores the previous directory afterward. If the directory doesn't
exist or can't be entered, an error is printed and the command is not run.

```
@dir=~/Projects/Hilbish git status
```

``` =html
<hr class="my-4">
```

### `runner`

Runs the command with a specific [runner mode](../runner-mode) instead of the
one currently set, just for that command.

```
@runner=lua print('hello from Lua')
```

``` =html
<hr class="my-4">
```

### `alias`

Setting this to `false` skips alias resolution, so the command runs exactly
as typed even if it matches a defined alias. The absence of this modifier
is treated as `true`.

```
@alias=false ss
```

``` =html
<hr class="my-4">
```
