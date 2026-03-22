---
title: Nature
layout: doc
weight: -90
menu: docs
---

A bit after creation, we have the outside nature. Little plants, seeds,
growing to their final phase: a full plant. A lot of Hilbish itself is
written in Go, but there are parts made in Lua, being most builtins
(`doc`, `cd`, cdr), completions, and other things.\
Hilbish's Lua core module is called `nature`.
It runs after Hilbish's Go core does.

# Nature Modules

Currently, `nature` provides 1 intended public module: `nature.dirs`.\
It is a simple API for managing recent directories and old
current working directory.
