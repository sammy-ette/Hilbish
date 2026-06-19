---@meta

---@class fs
---@field pathSep any
---@field abs fun(path: string): string
---@field basename fun(path: string): string
---@field cd fun(dir: string)
---@field dir fun(path: string): string
---@field executable fun(path: string): boolean
---@field glob fun(pattern: string): table
---@field join fun(...: string): string
---@field mkdir fun(name: string, recursive: boolean)
---@field pipe fun(): file*, file*
---@field readdir fun(dir: string): table
---@field stat fun(path: string): table
local fs = {}

return fs
