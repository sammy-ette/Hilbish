---
title: Keybindings
description: Configure how Hilbish responds to keyboard input
layout: doc
menu:
  docs:
    parent: "Features"
---

Readline keybindings controls how Hilbish interpretskeyboard input.
Hilbish exposes its readline editor instance through `hilbish.editor` so you can directly
modify and interact its keybinds and actions.

## Key Binding Concepts

Keybindings work in two layers:

1. **Keymap** maps keys (like `Ctrl-A` or arrow keys) to action names (`"cursor.beginning-of-line"`).
2. **Actions** are named and registered functions that perform those actions.

You can rebind keys to different actions, create new actions, or override builtin actions.

## Binding Keys

Use `bindKey(key, action)` to bind a key sequence to an action:

```lua
-- Bind to a built-in action
hilbish.editor:bindKey("Ctrl-X", "cursor.beginning-of-line")

-- Bind to a Lua function
hilbish.editor:bindKey("Ctrl-Y", function()
  print("Custom action!")
end)
```

## Registering Custom Actions

For reusable custom actions, use `addAction(name, fn)`:

```lua
hilbish.editor:addAction("my.custom.action", function()
  hilbish.editor:insert("Hello, world!")
end)

hilbish.editor:bindKey("Ctrl-H", "my.custom.action")
```

## Discoverability

You can query the current bindings and available actions with a simple function call:

```lua
local bindings = hilbish.editor:getBindings()
for key, action in pairs(bindings) do
  print(key .. ' -> '.. action)
end
```

This is useful for writing completion handlers, status line integrations, or debugging keybinding issues.

## Vim Mode

To switch to Vim mode, use:

```lua
hilbish.editor:setInputMode("vim")
```

Vim mode follows the same default keybindings listed for insert mode.
In normal mode Vim-specific motions are handled separately and can be customized
through the `setViModeCallback` and `setViActionCallback` methods.

See the [readline API documentation](/docs/api/readline) for more details on all available methods.

## Default Keybindings

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-C` | `cancel` | Cancels the current input. |
| `Ctrl-D` | `delete.char` | Deletes the character under the cursor; on empty line, sends EOF. |
| `Ctrl-L` | `screen.clear` | Clears the screen and redraws the prompt. |
| `Ctrl-U` | `delete.to-beginning` | Deletes all text from cursor to the beginning of the line. |
| `Ctrl-K` | `delete.to-end` | Deletes all text from cursor to the end of the line. |
| `Backspace` | `backspace` | You know what backspace does right? |
| `Ctrl-W` | `delete.kill-word-backward` | Deletes the word before the cursor. |
| `Ctrl-Y` | `register.yank` | Pastes the content of the last deleted text. |
| `Ctrl-E` | `cursor.end-of-line` | Moves the cursor to the end of the line. |
| `Ctrl-A` | `cursor.beginning-of-line` | Moves the cursor to the beginning of the line. |
| `Ctrl-R` | `history.search` | Begins history search. |
| `Tab` | `completion.toggle` | Toggles command completion. |
| `Ctrl-F` | `completion.search` | Searches completions. |
| `Ctrl-G` | `search.cancel` | Cancels the current search mode. |
| `Ctrl-_` | `undo` | Undoes the last edit. |
| `Enter` or `Ctrl-J` | `submit` | Submits the current line. |
| `Escape` | `escape` | Escapes from completion or search mode. |
| `Shift-Tab` | `completion.prev` | Moves to the previous completion. |
| `Up Arrow` | `history.prev` | Moves to the previous history entry. |
| `Down Arrow` | `history.next` | Moves to the next history entry. |
| `Right Arrow` | `cursor.forward` | Moves the cursor one character to the right. |
| `Left Arrow` | `cursor.backward` | Moves the cursor one character to the left. |
| `Alt-"` | `register.show` | Shows available registers. |
| `Ctrl-Left` | `cursor.move-word-backward` | Moves the cursor one word to the left. |
| `Ctrl-Right` | `cursor.move-word-forward` | Moves the cursor one word to the right. |
| `Alt-R` | `history.search-alt` | Begins alternative history search. |
| `Delete` | `delete.char-seq` | Deletes the character under the cursor. |
| `Home` | `cursor.beginning-of-line-seq` | Moves the cursor to the beginning of the line. |
| `End` | `cursor.end-of-line-seq` | Moves the cursor to the end of the line. |
| `Alt-B` | `cursor.word-backward` | Moves the cursor to the beginning of the current word. |
| `Alt-F` | `cursor.word-forward` | Moves the cursor to the end of the next word. |
| `Alt-Backspace` | `delete.word-backward` | Deletes the word before the cursor. |
| `Ctrl-Delete` | `delete.word-forward` | Deletes the word after the cursor. |
| `Page-Up` | `history.prev` | Moves to the previous history entry. |
| `Page-Down` | `history.next` | Moves to the next history entry. |

## Available Actions

All built-in actions are listed below. You can bind any key to any of these actions.

| Action | Description |
|--------|-------------|
| `cancel` | Cancels the current input, sending an interrupt signal. |
| `delete.char` | Deletes the character under the cursor; on empty line, sends EOF. |
| `screen.clear` | Clears the screen and redraws the prompt. |
| `delete.to-beginning` | Deletes all text from cursor to the beginning of the line. |
| `delete.to-end` | Deletes all text from cursor to the end of the line. |
| `backspace` | Surely you know what this does! |
| `delete.kill-word-backward` | Deletes the word before the cursor. |
| `register.yank` | Pastes the content of the last deleted text. |
| `cursor.end-of-line` | Moves the cursor to the end of the line. |
| `cursor.beginning-of-line` | Moves the cursor to the beginning of the line. |
| `history.search` | Begins reverse history search. |
| `completion.toggle` | Toggles command completion. |
| `completion.search` | Searches completions. |
| `search.cancel` | Cancels the current search mode. |
| `undo` | Undoes the last edit. |
| `submit` | Submits the current line for execution. |
| `escape` | Escapes from completion or search mode. |
| `completion.prev` | Moves to the previous completion. |
| `history.prev` | Moves to the previous history entry. |
| `history.next` | Moves to the next history entry. |
| `cursor.forward` | Moves the cursor one character to the right. |
| `cursor.backward` | Moves the cursor one character to the left. |
| `register.show` | Shows available registers. |
| `cursor.move-word-backward` | Moves the cursor one word to the left. |
| `cursor.move-word-forward` | Moves the cursor one word to the right. |
| `cursor.word-backward` | Moves the cursor to the beginning of the current word. |
| `cursor.word-forward` | Moves the cursor to the end of the next word. |
| `history.search-alt` | Begins alternative history search. |
| `delete.char-seq` | Deletes the character under the cursor. |
| `cursor.beginning-of-line-seq` | Moves the cursor to the beginning of the line. |
| `cursor.end-of-line-seq` | Moves the cursor to the end of the line. |
| `delete.word-backward` | Deletes the word before the cursor. |
| `delete.word-forward` | Deletes the word after the cursor. |
