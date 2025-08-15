---
title: Getting Started
layout: doc
weight: -10
menu: docs
---

To start Hilbish, open a terminal. If Hilbish has been installed and is not the
default shell, you can simply run `hilbish` to start it. This will launch
a normal interactive session.
To exit, you can either run the `exit` command or hit Ctrl+D.

# Setting as Default

## Login shell

There are a few ways to make Hilbish your default shell. A simple way is 
to make it your user/login shell.

{{< warning `It is not recommended to set Hilbish as your login shell. That
is expected to be a POSIX compliant shell, which Hilbish is not. Though if
you still decide to do it, there will just be a few variables missing in
your environment` >}}

To do that, simply run `chsh -s /usr/bin/hilbish`.

Some distros (namely Fedora) might have `lchsh` instead, which is used like `lchsh <user>`.
When prompted, you can put the path for Hilbish.

## Default with terminal

The simpler way is to set the default shell for your terminal. The way of 
doing this depends on how your terminal settings are configured.

## Run after login shell

Some shells (like zsh) have an rc file, like `.zlogin`, which is ran when the shell session
is a login shell. In that file, you can run Hilbish with this command: `exec hilbish -S -l`

This will replace the shell with Hilbish, set $SHELL to Hilbish and launch it as a login shell.

# Configuration

Once installation and setup has been done, you can then configure Hilbish.
It is configured and scripted via Lua, so the config file is a Lua file.
You can use any pure Lua library to do whatever you want.\
 \
Hilbish's sample configuration is usually located in `hilbish.dataDir .. '/.hilbishrc.lua'`.
You can print that path via Lua to see what it is: `print(hilbish.dataDir .. '/.hilbishrc.lua')`.\
 \
As an example, it will usually will result in `/usr/share/hilbish/.hilbishrc.lua` on Linux.\
 \
To edit your user configuration, you can copy that file to `hilbish.userDir.config .. '/hilbish/init.lua'`, which follows XDG on Linux and MacOS, and is located in %APPDATA% on Windows.\
 \
As the directory is usually `~/.config` on Linux, you can run this command to copy it:  
`cp /usr/share/hilbish/.hilbishrc.lua ~/.config/hilbish/init.lua`. Now we can get to customization!\
 \
If we closely examine a small snippet of the default config:

```lua
-- Default Hilbish config
-- .. with some omitted code .. --

local function doPrompt(fail)
	hilbish.prompt(lunacolors.format(
		'{blue}%u {cyan}%d ' .. (fail and '{red}' or '{green}') .. '∆ '
	))
end

doPrompt()

bait.catch('command.exit', function(code)
	doPrompt(code ~= 0)
end)
```
\
 
We see a whopping **three** Hilbish libraries being used in this part of code.
First is of course, named after the shell itself, [`hilbish`](../api/hilbish). This is kind of a
"catch-all" namespace for functions that directly related to shell functionality/settings.\
 \
And as we can see, the [hilbish.prompt](../api/hilbish/#prompt) function is used
to change our prompt. Change our prompt to what, exactly?\
The doc for the function states that the verbs `%u` and `%d` are used for username and current directory of the shell, respectively.\
We wrap this in the [`lunacolors.format`](../lunacolors) function, to give
our prompt some nice color.\
 \
 
But you might have also noticed that this is in the `doPrompt` function, which is called once,
and then used again in a [bait](../api/bait) hook. Specifically, the `command.exit` hook,
which is called after a command exits, so when it finishes running.
