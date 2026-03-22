---
title: Frequently Asked Questions
layout: doc
weight: -20
menu: docs
---

# Is Hilbish POSIX compliant?

No, it is not. POSIX compliance is a non-goal. Perhaps in the future,
someone would be able to write a native plugin to support shell scripting
(which would be against it's main goal, but ....)

# Windows Support?

It compiles for Windows (CI ensures it does), but otherwise it is not
directly supported. If you'd like to improve this situation,
checkout [the discussion](https://github.com/sammy-ette/Hilbishsh/discussions/165).

# Why?

Hilbish emerged from the desire of a Lua configured shell.
It was the initial reason that it was created, but now it's more:
to be hyper extensible, simpler and more user friendly.

# Does it have "autocompletion" or "tab completion"

Of course! This is a modern shell. Hilbish provides a way for users
to write tab completion for any command and/or the whole shell.
Inline hinting and syntax highlighting are also available.
