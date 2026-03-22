---
title: Actions
layout: doc
weight: -80
menu: 
  docs:
    parent: "Vim Mode"
---

Vim actions are essentially just when a user uses a Vim keybind.
Things like yanking and pasting are Vim actions.
This is not an "offical Vim thing," just a Hilbish thing.\
 \
The `hilbish.vimAction` hook is thrown whenever a Vim action occurs.
It passes 2 arguments: the action name, and an array (table) of args
relating to it.\
 \
Here is documentation for what the table of args will hold for an
appropriate Vim action.

- `yank`: register, yankedText
The first argument for the yank action is the register yankedText goes to.

- `paste`: register, pastedText
The first argument for the paste action is the register pastedText is taken from.
