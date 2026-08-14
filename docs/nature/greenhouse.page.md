---
title: Module greenhouse.page
description: page object for greenhouse
layout: doc
menu:
  docs:
    parent: "Nature"
---

## Introduction

The `greenhouse.page` module provides the Page class, which represents a single page that can be displayed
in the Greenhouse pager. Pages are used to organize and display content in a scrollable, navigable format.

## Types

---

## Page

A Page represents a single document or view that can be displayed in the Greenhouse pager.

### Constructor

#### new

:::signature
```lua
Page:new(title, text) -> Page
```
:::

Creates a new Page with the given title and text content.

#### Parameters

:::params
`string` _title_  
The title of the page, displayed at the top.

`string` _text_  
The text content of the page. Lines are split by newlines.

:::

#### Returns

:::returns
`Page`  
The newly created page object.

:::

### Methods

---

#### setText

:::signature
```lua
page:setText(text)
```
:::

Sets or updates the text content of the page. The text is split into lines by newlines.

#### Parameters

:::params
`string` _text_  
The new text content for the page.

:::

---

#### setTitle

:::signature
```lua
page:setTitle(title)
```
:::

Sets or updates the title of the page.

#### Parameters

:::params
`string` _title_  
The new title for the page.

:::

---

#### dynamic

:::signature
```lua
page:dynamic(initializer)
```
:::

Marks the page as lazy-loaded with a dynamic initializer function.
This is used for pages that should load their content on demand.
The initializer function will be called when the page needs to be loaded.

#### Parameters

:::params
`function` _initializer_  
A function that will be called to initialize the page content.

:::

---

#### initialize

:::signature
```lua
page:initialize()
```
:::

Initializes a lazy-loaded page by calling its initializer function.
This is typically called internally by Greenhouse when the page is first displayed.

### Object Properties

- `title`: The title string of the page.
- `lines`: A table of text lines that make up the page content.
- `lazy`: Whether the page is lazy-loaded (boolean).
- `loaded`: Whether the page has been initialized (boolean).
- `children`: A table of child pages or related pages.
