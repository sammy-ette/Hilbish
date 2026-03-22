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
|||
|----|----|
|<a href="#renderInfoBlock">renderInfoBlock(type, text)</a>|Renders an info block. An info block is a block of text with|
|<a href="#renderCodeBlock">renderCodeBlock(text)</a>|Assembles and renders a code block. This returns|
|<a href="#highlight">highlight(text)</a>|Performs basic Lua code highlighting.|
<hr>
<div id='highlight'>
<h4 class='heading'>
doc.highlight(text)
<a href="#highlight" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Performs basic Lua code highlighting.
#### Parameters
`text` **`string`**  
 Code/text to do highlighting on.

</div>

<hr>
<div id='renderCodeBlock'>
<h4 class='heading'>
doc.renderCodeBlock(text)
<a href="#renderCodeBlock" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Assembles and renders a code block. This returns
the supplied text based on the number of command line columns,
and styles it to resemble a code block.
#### Parameters
`text` **`string`**  


</div>

<hr>
<div id='renderInfoBlock'>
<h4 class='heading'>
doc.renderInfoBlock(type, text)
<a href="#renderInfoBlock" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Renders an info block. An info block is a block of text with
an icon and styled text block.
#### Parameters
`type` **`string`**  
 Type of info block. The only one specially styled is the `warning`.

`text` **`string`**  


</div>

