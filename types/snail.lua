---@meta

---@class Snail
local Snail = {}

---@param path string
function Snail:dir(path) end

---@param command string
---@param streams? table
---@return table
function Snail:run(command, streams) end

---@class snail
---@field new fun(): Snail
---@field validate fun(input: string)
local snail = {}

return snail
