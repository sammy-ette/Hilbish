---
title: Notification
description: Get notified of shell actions.
layout: doc
menu: 
  docs:
    parent: "Features"
---

Hilbish features a simple notification system which can be
used by other plugins and parts of the shell to notify the user
of various actions. This is used via the `hilbish.message` interface.\
 \
A `message` is defined as a table with the following properties:

- `icon`: A unicode/emoji icon for the notification.
- `title`: The title of the message
- `text`: Message text/body
- `channel`: The source of the message. This should be a unique and easily readable text identifier.
- `summary`: A short summary of the notification and message. If this is not present and you are using this to display messages, you should take part of the `text` instead.

 \
The `hilbish.message` interface provides the following functions:

- `send(message)`: Sends a message and emits the `hilbish.notification` signal. DO NOT emit the `hilbish.notification` signal directly, or the message will not be stored by the message handler.
- `read(idx)`: Marks message at `idx` as read.
- `delete(idx)`: Removes message at `idx`.
- `readAll()`: Marks all messages as read.
- `clear()`: Deletes all messages.

 \
There are a few simple use cases of this notification/messaging system.
It could also be used as some "inter-shell" messaging system (???) but
is intended to display to users.\
 \
An example is notifying users of completed jobs/commands ran in the background.
Any Hilbish-native command (think the upcoming Greenhouse pager) can display
it.
