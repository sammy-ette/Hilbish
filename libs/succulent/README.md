# Succulent
> 🌿 Lua functions and utilities specifically for Hilbish.

Succulent is a small and simple pure Lua library which adds extra useful functions
to the standard library. Succulent is made for [Hilbish](https://github.com/Rosettea/Hilbish) specifically.

# Usage
If on Hilbish, Succulent should already be added globally. If not, simply `require` it
and all functions will be added.

# Documentation
#### string
`split(str, delim)` > Divides a string into a table of substrings separated by `delim`.

#### table
`filter(t, fn)` > Returns a table with all elements of `t` that pass the filter provided by `fn`.  
`map(t, fn)` > Returns a new table with the results of calling `fn` on the elements in `t`.

# License
MIT

