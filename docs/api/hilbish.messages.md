---
title: Module hilbish.messages
description: simplistic message passing
layout: doc
menu:
  docs:
    parent: "API"
---

## Introduction

The messages interface defines a way for Hilbish-integrated commands,
user config and other tasks to send notifications to alert the user.


A message starts out unread when it is sent with `send`. It stays
unread until it is explicitly marked with `read` (by index) or
`readAll` (every message at once). `unreadCount` and `all` reflect
this state, so a prompt or statusline can show how many messages are
waiting to be seen.


```lua
hilbish.messages.send{
	title = 'Build finished',
	text = 'go build exited with code 0',
	channel = 'build',
	icon = '✅'
}


print(hilbish.messages.unreadCount()) -- -> 1


for idx, msg in pairs(hilbish.messages.all()) do
	hilbish.messages.read(idx)
end
print(hilbish.messages.unreadCount()) -- -> 0
```

## Functions

:::funclist
- [`hilbish.messages.all() -> table`](#all): Returns all messages as a table keyed by their index.
- [`hilbish.messages.clear()`](#clear): Deletes all messages.
- [`hilbish.messages.delete(idx)`](#delete): Deletes the message at `idx`. Errors if the index is invalid.
- [`hilbish.messages.read(idx)`](#read): Marks a message at `idx` as read.
- [`hilbish.messages.readAll()`](#readAll): Marks all messages as read.
- [`hilbish.messages.send(message)`](#send): Sends a notification message and emits the `hilbish.notification` signal.
- [`hilbish.messages.unreadCount() -> integer`](#unreadCount): Returns the count of unread messages.

:::

---

#### all

:::signature
```lua
hilbish.messages.all() -> table
```
:::

Returns all messages as a table keyed by their index.  

#### Returns

:::returns
`table`  
All stored messages, keyed by message index.

:::



---

#### clear

:::signature
```lua
hilbish.messages.clear()
```
:::

Deletes all messages.  



---

#### delete

:::signature
```lua
hilbish.messages.delete(idx)
```
:::

Deletes the message at `idx`. Errors if the index is invalid.  

#### Parameters

:::params
`number` _idx_  
Index of the message to delete.

:::



---

#### read

:::signature
```lua
hilbish.messages.read(idx)
```
:::

Marks a message at `idx` as read.  

#### Parameters

:::params
`number` _idx_  
Index of the message to mark as read.

:::



---

#### readAll

:::signature
```lua
hilbish.messages.readAll()
```
:::

Marks all messages as read.  



---

#### send

:::signature
```lua
hilbish.messages.send(message)
```
:::

Sends a notification message and emits the `hilbish.notification` signal.  
Do *not* emit the `hilbish.notification` signal directly.  

#### Parameters

:::params
`hilbish.message` _message_  

:::



---

#### unreadCount

:::signature
```lua
hilbish.messages.unreadCount() -> integer
```
:::

Returns the count of unread messages.  

#### Returns

:::returns
`integer`  
Number of messages that have not been marked as read.

:::



## Types

---

## hilbish.message

Represents a Hilbish message.

## Object Properties

- `string?` `icon`: Unicode (preferably standard emoji) icon for the message notification.
- `string` `title`: Title of the message (like an email subject).
- `string` `text`: Contents of the message.
- `string` `channel`: Short identifier of the message. `hilbish` and `hilbish.*` is preserved for internal Hilbish messages.
- `string` `summary`: A short summary of the message.
- `boolean?` `read`: Whether the full message has been read or not.


### Methods

