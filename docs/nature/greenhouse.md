---
title: Module greenhouse
description: Hilbish's environmentally friendly pager for long text
layout: doc
menu:
  docs:
    parent: "Nature"
---

## Introduction

Greenhouse is a simple text scrolling handler (pager) for terminal programs.
It can be set a specific region to do its scrolling and paging job,
then the user can draw whatever outside it. You may use this as some kind
of terminal UI library.

## Functions

:::funclist
- [`greenhouse.new(sink) -> nature.object`](#new): Creates a new Greenhouse pager.
- [`greenhouse.render()`](#render): This should be overloaded to render extra UI around Greenhouse.
- [`greenhouse.resize()`](#resize): Resizes Greenhouse to fit the terminal's size.
- [`greenhouse.scroll(direction, opts)`](#scroll): Scrolls the currently active page by one line, or a page height if specified.
- [`greenhouse.updateSpecial()`](#updateSpecial): This function will be called when the special page

:::

---

#### new

:::signature
```lua
greenhouse.new(sink) -> nature.object
```
:::

Creates a new Greenhouse pager.  

#### Parameters

:::params
`hilbish.sink` _sink_  

:::

#### Returns

:::returns
`nature.object`  

:::



---

#### render

:::signature
```lua
greenhouse.render()
```
:::

This should be overloaded to render extra UI around Greenhouse.  



---

#### resize

:::signature
```lua
greenhouse.resize()
```
:::

Resizes Greenhouse to fit the terminal's size.  



---

#### scroll

:::signature
```lua
greenhouse.scroll(direction, opts)
```
:::

Scrolls the currently active page by one line, or a page height if specified.  

#### Parameters

:::params
`string` _direction_  
Either `up` or `down`

`table` _opts_  

:::tparams
`boolean` _page_  
Whether the scroll amount should be the page height.

:::

:::



---

#### updateSpecial

:::signature
```lua
greenhouse.updateSpecial()
```
:::

This function will be called when the special page  
is on and needs to be updated.  



