---
title: Hilbish
description:
layout: doc
menu:
  docs:
    parent: "Signals"
---

## hilbish.exit

Sent when Hilbish is going to exit.

#### Variables

This signal returns no variables.

``` =html
<hr class="my-4">
```

## hilbish.vimMode

Sent when the Vim mode of Hilbish is changed (like from insert to normal mode).
This can be used to change the prompt and notify based on Vim mode.

#### Variables

`string` _modeName_
The mode that has been set.
Can be these values: `insert`, `normal`, `delete` or `replace`

``` =html
<hr class="my-4">
```

## hilbish.cancel

Sent when the user cancels their command input with Ctrl-C

#### Variables

This signal returns no variables.

``` =html
<hr class="my-4">
```

## hilbish.notification

Thrown when a [notification](../../features/notifications) is sent.

#### Variables

`table` _notification_
The notification. The properties are defined in the link above.

``` =html
<hr class="my-4">
```

## hilbish.cd

Sent when the current directory of the shell is changed (via interactive means.)
~~If you are implementing a custom command that changes the directory of the shell, you must throw this hook manually for correctness.~~ Since 3.0, `hilbish.cd` is thrown when `fs.cd` is called.

#### Variables

`string` _path_
Absolute path of the directory that was changed to.

`string` _oldPath_
Absolute path of the directory Hilbish *was* in.

``` =html
<hr class="my-4">
```

## hilbish.vimAction

Sent when the user does a "vim action," being something like yanking or pasting text.
See `doc vim-mode actions` for more info.

#### Variables

`string` _actionName_
Absolute path of the directory that was changed to.

`table` _args_
Table of args relating to the Vim action.

``` =html
<hr class="my-4">
```
