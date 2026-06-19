---@meta

---@class bait
---@field catch fun(name: string, cb: fun(...: any))
---@field catchOnce fun(name: string, cb: fun(...: any))
---@field hooks fun(name: string): table<fun(...: any)>
---@field release fun(name: string, catcher: fun(...: any))
---@field throw fun(name: string, ...: any)
local bait = {}

return bait
