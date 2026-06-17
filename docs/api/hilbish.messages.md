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
user config and other tasks to send notifications to alert the user.z
The `hilbish.message` type is a table with the following keys:
`title` (string): A title for the message notification.
`text` (string): The contents of the message.
`channel` (string): States the origin of the message, `hilbish.*` is reserved for Hilbish tasks.
`summary` (string): A short summary of the `text`.
`icon` (string): Unicode (preferably standard emoji) icon for the message notification
`read` (boolean): Whether the full message has been read or not.

## Functions

- [`hilbish.messages.all() -> table<hilbish.message>`](#all): Returns all messages.
- [`hilbish.messages.clear()`](#clear): Deletes all messages.
- [`hilbish.messages.delete(idx)`](#delete): Deletes the message at `idx`.
- [`hilbish.messages.read(idx)`](#read): Marks a message at `idx` as read.
- [`hilbish.messages.readAll()`](#readAll): Marks all messages as read.
- [`hilbish.messages.send(message)`](#send): Sends a message.
- [`hilbish.messages.unreadCount() -> integer`](#unreadCount): Returns the amount of unread messages.

---

#### all

hilbish.messages.all() -> table<hilbish.message>

Returns all messages.  

#### Parameters

This function has no parameters.  


---

#### clear

hilbish.messages.clear()

Deletes all messages.  

#### Parameters

This function has no parameters.  


---

#### delete

hilbish.messages.delete(idx)

Deletes the message at `idx`.  

#### Parameters

`number` _idx_  




---

#### read

hilbish.messages.read(idx)

Marks a message at `idx` as read.  

#### Parameters

`number` _idx_  




---

#### readAll

hilbish.messages.readAll()

Marks all messages as read.  

#### Parameters

This function has no parameters.  


---

#### send

hilbish.messages.send(message)

Sends a message.  

#### Parameters

`hilbish.message` _message_  




---

#### unreadCount

hilbish.messages.unreadCount() -> integer

Returns the amount of unread messages.  

#### Parameters

This function has no parameters.  


