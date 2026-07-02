---
title: Module doc
description: command-line doc rendering
layout: doc
menu:
  docs:
    parent: "Nature"
---

## Introduction

The doc module contains a small set of functions
used by the Greenhouse pager to render parts of the documentation pages.
This is only documented for the sake of it. It's only intended use
is by the Greenhouse pager.

## Functions

:::funclist
- [`doc.highlight(text)`](#highlight): Performs basic Lua code highlighting.
- [`doc.renderCodeBlock(text)`](#renderCodeBlock): Assembles and renders a code block. This returns
- [`doc.renderInfoBlock(type, text)`](#renderInfoBlock): Renders an info block. An info block is a block of text with

:::

---

#### highlight

:::signature
```lua
doc.highlight(text)
```
:::

Performs basic Lua code highlighting.  

#### Parameters

:::params
`string` _text_  
Code/text to do highlighting on.

:::



---

#### renderCodeBlock

:::signature
```lua
doc.renderCodeBlock(text)
```
:::

Assembles and renders a code block. This returns  
the supplied text based on the number of command line columns,  
and styles it to resemble a code block.  

#### Parameters

:::params
`string` _text_  

:::



---

#### renderInfoBlock

:::signature
```lua
doc.renderInfoBlock(type, text)
```
:::

Renders an info block. An info block is a block of text with  
an icon and styled text block.  

#### Parameters

:::params
`string` _type_  
Type of info block. The only one specially styled is the `warning`.

`string` _text_  

:::



