---
title: Module hilbish.os
description: operating system info
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

Provides simple text information properties about the current operating system.
This mainly includes the name and version.


This is commonly used to branch config behavior per platform, for example
picking a different package manager completer or prompt icon:


```lua
if hilbish.os.family == 'windows' then
	hilbish.prompt '%u@%h %d>'
else
	hilbish.prompt '%u@%h %d $'
end
```

## Static module fields

:::fieldlist
- `string` `family`: Family name of the current OS
- `string` `name`: Pretty name of the current OS
- `string` `version`: Version of the current OS

:::

