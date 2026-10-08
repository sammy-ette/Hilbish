---@meta

---@class commander
---@field deregister fun(name: string)
---@field get fun(name: string): fun(args:table,sinks:table<string,Sink>):number?
---@field register fun(name: string, cb: fun(args:table,sinks:table<string,Sink>):number?)
---@field registry fun(): table
local commander = {}

return commander
