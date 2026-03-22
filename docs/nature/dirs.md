---
title: Module dirs
description: internal directory management
layout: doc
menu:
  docs:
    parent: "Nature"
---


## Introduction
The dirs module defines a small set of functions to store and manage
directories.

## Functions
|||
|----|----|
|<a href="#setOld">setOld(d)</a>|Sets the old directory string.|
|<a href="#recent">recent(idx)</a>|Get entry from recent directories list based on index.|
|<a href="#push">push(dir)</a>|Add `dir` to the recent directories list.|
|<a href="#pop">pop(num)</a>|Remove the specified amount of dirs from the recent directories list.|
|<a href="#peak">peak(num)</a>|Look at `num` amount of recent directories, starting from the latest.|
<hr>
<div id='peak'>
<h4 class='heading'>
dirs.peak(num)
<a href="#peak" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Look at `num` amount of recent directories, starting from the latest.
This returns  a table of recent directories, up to the `num` amount.
#### Parameters
`num` **`number`**  


</div>

<hr>
<div id='pop'>
<h4 class='heading'>
dirs.pop(num)
<a href="#pop" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Remove the specified amount of dirs from the recent directories list.
#### Parameters
`num` **`number`**  


</div>

<hr>
<div id='push'>
<h4 class='heading'>
dirs.push(dir)
<a href="#push" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Add `dir` to the recent directories list.
#### Parameters
`dir` **`string`**  


</div>

<hr>
<div id='recent'>
<h4 class='heading'>
dirs.recent(idx)
<a href="#recent" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Get entry from recent directories list based on index.
#### Parameters
`idx` **`number`**  


</div>

<hr>
<div id='setOld'>
<h4 class='heading'>
dirs.setOld(d)
<a href="#setOld" class='heading-link'>
	<i class="fas fa-paperclip"></i>
</a>
</h4>

Sets the old directory string.
#### Parameters
`d` **`string`**  


</div>

