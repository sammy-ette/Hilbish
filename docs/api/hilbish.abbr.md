---
title: Module hilbish.abbr
description: command line abbreviations
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The abbr module manages Hilbish abbreviations. These are words that can be replaced
with longer command line strings when entered.
As an example, `git push` can be abbreviated to `gp`. When the user types
`gp` into the command line, after hitting space or enter, it will expand to `git push`.
Abbreviations can be used as an alternative to aliases. They are saved entirely in the history
Instead of the aliased form of the same command.

## Functions

- [`hilbish.abbr.add(abbr, expanded, opts)`](#add): Adds an abbreviation. The `abbr` is the abbreviation itself,
- [`hilbish.abbr.remove(abbr)`](#remove): Removes the named `abbr`.

---

#### add

hilbish.abbr.add(abbr, expanded, opts)

Adds an abbreviation. The `abbr` is the abbreviation itself,  
while `expanded` is what the abbreviation should expand to.  
It can be either a function or a string. If it is a function, it will expand to what  
the function returns.  
`opts` is a table that accepts 1 key: `anywhere`.  
`opts.anywhere` defines whether the abbr expands anywhere in the command line or not,  
whereas the default behavior is only at the beginning of the line  

#### Parameters

`string` _abbr_  


`string|function` _expanded_  


`table` _opts?_  




---

#### remove

hilbish.abbr.remove(abbr)

Removes the named `abbr`.  

#### Parameters

`string` _abbr_  




