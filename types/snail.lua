---@meta

---@class Snail
local Snail = {}

---@param path string
function Snail:dir(path) end

---@param command string
---@param streams? table
---@return table result
function Snail:run(command, streams) end

---@class snail
---@field new fun(): Snail
---@field validate fun(input: string): boolean
local snail = {}

return snail
