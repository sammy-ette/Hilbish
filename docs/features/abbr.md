---
title: Abbreviations
description: Words that expand in-place on the command line.
layout: doc
weight: -15
menu: 
  docs:
    parent: "Features"
---

Abbreviations are similar to [aliases](../aliases) but work differently: when
you type an abbreviation and press space or enter, the word is replaced
in-place on the command line with the full text before anything runs. The
expanded form is what gets saved to history.

This is the main practical difference from aliases: with aliases the short name
goes to history; with abbreviations the expanded form does. If you care about
your history containing readable, full commands rather than short codes,
abbreviations are the better choice.

By default abbreviations only expand at the beginning of the line. They can
optionally be set to expand anywhere on the line.

Abbreviations are defined in your config file. See the
[hilbish.abbr API](../../api/hilbish/hilbish.abbr) for how to add and remove them.
