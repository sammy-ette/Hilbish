-- @module hilbish.messages
-- simplistic message passing
-- The messages interface defines a way for Hilbish-integrated commands,
-- user config and other tasks to send notifications to alert the user.z
-- The `hilbish.message` type is a table with the following keys:
-- `title` (string): A title for the message notification.
-- `text` (string): The contents of the message.
-- `channel` (string): States the origin of the message, `hilbish.*` is reserved for Hilbish tasks.
-- `summary` (string): A short summary of the `text`.
-- `icon` (string): Unicode (preferably standard emoji) icon for the message notification
-- `read` (boolean): Whether the full message has been read or not.
local bait = require 'bait'
local commander = require 'commander'
local lunacolors = require 'lunacolors'

local M = {}
local counter = 0
local unread = 0
M._messages = {}
M.icons = {
	INFO = '',
	SUCCESS = '',
	WARN = '',
	ERROR = ''
}

hilbish.messages = {}

--- Represents a Hilbish message.
--- @class hilbish.message
--- @field icon string Unicode (preferably standard emoji) icon for the message notification.
--- @field title string Title of the message (like an email subject).
--- @field text string Contents of the message.
--- @field channel string Short identifier of the message. `hilbish` and `hilbish.*` is preserved for internal Hilbish messages.
--- @field summary string A short summary of the message.
--- @field read boolean Whether the full message has been read or not.

function expect(tbl, field)
	if not tbl[field] or tbl[field] == '' then
		error(string.format('expected field %s in message'))
	end
end

--- Sends a message.
--- @param message hilbish.message
function hilbish.messages.send(message)
	expect(message, 'text')
	expect(message, 'title')
	counter = counter + 1
	unread = unread + 1
	message.index = counter
	message.read = false

	M._messages[message.index] = message
	bait.throw('hilbish.notification', message)
end

--- Marks a message at `idx` as read.
--- @param idx number
function hilbish.messages.read(idx)
	local msg = M._messages[idx]
	if msg then 
		M._messages[idx].read = true
		unread = unread - 1
	end
end

--- Marks all messages as read.
function hilbish.messages.readAll()
	for _, msg in ipairs(hilbish.messages.all()) do
		hilbish.messages.read(msg.index)
	end
end

--- Returns the amount of unread messages.
function hilbish.messages.unreadCount()
	return unread
end

--- Deletes the message at `idx`.
--- @param idx number
function hilbish.messages.delete(idx)
	local msg = M._messages[idx]
	if not msg then
		error(string.format('invalid message index %d', idx or -1))
	end

	M._messages[idx] = nil
end

--- Deletes all messages.
function hilbish.messages.clear()
	for _, msg in ipairs(hilbish.messages.all()) do
		hilbish.messages.delete(msg.index)
	end
end

--- Returns all messages.
function hilbish.messages.all()
	return M._messages
end

return M
