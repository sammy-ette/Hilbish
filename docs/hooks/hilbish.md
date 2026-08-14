---
title: Hilbish
description:
layout: doc
menu:
  docs:
    parent: "Signals"
---

## hilbish.cancel

Sent when the user cancels their command input with Ctrl-C.

#### Variables

This signal has no variables.

``` =html
<hr class="my-4">
```

## hilbish.cd

Sent when the current directory of the shell is changed.
Since 3.0, hilbish.cd is thrown when fs.cd is called.

#### Variables

`string` _path_  
Absolute path of the directory that was changed to.

`string` _oldPath_  
Absolute path of the directory Hilbish was in before the change.


``` =html
<hr class="my-4">
```

## hilbish.exit

Sent when Hilbish is going to exit.

#### Variables

This signal has no variables.

``` =html
<hr class="my-4">
```

## hilbish.init

Thrown once, right after the user's config file has finished loading
successfully, in interactive sessions.

#### Variables

This signal has no variables.

``` =html
<hr class="my-4">
```

## hilbish.notification

Thrown when a [notification](../features/notifications) is sent.

#### Variables

`table` _notification_  
The notification. See the notifications feature doc for its properties.


``` =html
<hr class="my-4">
```

## hilbish.rawInput

Thrown when the user has entered text in the prompt and pressed enter, before any processing.

#### Variables

`string` _input_  
The raw input the user typed.


``` =html
<hr class="my-4">
```

## hilbish.vimAction

Sent when the user does a "vim action," such as yanking or pasting text.
See `doc vim-mode actions` for a list of available actions.

#### Variables

`string` _actionName_  
The name of the Vim action that was done.

`table` _args_  
Table of args relating to the Vim action.


``` =html
<hr class="my-4">
```

## hilbish.vimMode

Sent when the Vim mode of Hilbish is changed (like from insert to normal mode).
This can be used to change the prompt and notify based on Vim mode.

#### Variables

`string` _modeName_  
The mode that has been set. Can be: `insert`, `normal`, `delete` or `replace`.

