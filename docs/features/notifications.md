---
title: Notifications
description: Get notified of shell actions.
layout: doc
menu: 
  docs:
    parent: "Features"
---

Hilbish has an internal notification system that shell features and plugins can
use to surface information to the user. A notification is a small message with
a title, body, icon, and channel (the source that sent it).

By itself, Hilbish does not display notifications in any special way — what
happens with a notification depends entirely on what your config does with it.
Common uses include:

- Showing an unread count badge in the prompt
- Printing a message when a background job finishes
- Alerting when a long-running command completes

The `notifyJobFinish` option (see [Options](../opts)) uses this system to print
a notification when a background job exits.

For sending notifications or reacting to them from your config, see the
[hilbish.messages API](../../api/hilbish/hilbish.messages) and the
[hilbish.notification signal](../../hooks/hilbish/#hilbish.notification).
