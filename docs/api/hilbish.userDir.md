---
title: Module hilbish.userDir
description: user-related directories
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

This interface just contains properties to know about certain user directories.
It is equivalent to XDG on Linux and gets the user's preferred directories
for configs and data.


This is mainly useful for locating files alongside your own config, since
Hilbish's own config lives at `fs.join(hilbish.userDir.config, 'hilbish')`:


```lua
local fs = require 'fs'
local myPluginDir = fs.join(hilbish.userDir.config, 'hilbish', 'myplugin')
```

## Static module fields

:::fieldlist
- `string` `config`: The user's config directory
- `string` `data`: The user's directory for program data

:::

