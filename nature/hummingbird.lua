-- @module hilbish.messages
-- simplistic message passing
-- The messages interface defines a way for Hilbish-integrated commands,
-- user config and other tasks to send notifications to alert the user.
-- <nl>
-- A message starts out unread when it is sent with `send`. It stays
-- unread until it is explicitly marked with `read` (by index) or
-- `readAll` (every message at once). `unreadCount` and `all` reflect
-- this state, so a prompt or statusline can show how many messages are
-- waiting to be seen.
-- <nl>
-- ```lua
-- hilbish.messages.send{
-- 	title = 'Build finished',
-- 	text = 'go build exited with code 0',
-- 	channel = 'build',
-- 	icon = '✅'
-- }
-- <nl>
-- print(hilbish.messages.unreadCount()) -- -> 1
-- <nl>
-- for idx, msg in pairs(hilbish.messages.all()) do
-- 	hilbish.messages.read(idx)
-- end
-- print(hilbish.messages.unreadCount()) -- -> 0
-- ```
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

---@diagnostic disable-next-line: missing-fields
hilbish.messages = {}

--- Represents a Hilbish message.
--- @class hilbish.message
--- @field icon? string Unicode (preferably standard emoji) icon for the message notification.
--- @field title string Title of the message (like an email subject).
--- @field text string Contents of the message.
--- @field channel string Short identifier of the message. `hilbish` and `hilbish.*` is preserved for internal Hilbish messages.
--- @field summary string A short summary of the message.
--- @field read? boolean Whether the full message has been read or not.

local function expect(tbl, field)
	if not tbl[field] or tbl[field] == '' then
		error(string.format('expected field %s in message', field))
	end
end

--- Sends a notification message and emits the `hilbish.notification` signal.
--- Do *not* emit the `hilbish.notification` signal directly.
--- @param message hilbish.message
function hilbish.messages.send(message)
	expect(message, 'text')
	expect(message, 'title')
	counter = counter + 1
	unread = unread + 1
	---@diagnostic disable-next-line: inject-field
	message.index = counter
	message.read = false

	M._messages[message.index] = message
	bait.throw('hilbish.notification', message) -- see nature/hooks.lua
end

--- Marks a message at `idx` as read.
--- @param idx number Index of the message to mark as read.
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

--- Returns the count of unread messages.
--- @return integer count Number of messages that have not been marked as read.
function hilbish.messages.unreadCount()
	return unread
end

--- Deletes the message at `idx`. Errors if the index is invalid.
--- @param idx number Index of the message to delete.
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

--- Returns all messages as a table keyed by their index.
--- @return table messages All stored messages, keyed by message index.
function hilbish.messages.all()
	return M._messages
end

return M
