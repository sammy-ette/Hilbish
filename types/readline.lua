---@meta

---@class Readline
local Readline = {}

---@param amount number
function Readline:DeleteByAmount(amount) end

---@return string line
function Readline:GetLine() end

---@param register string
---@return string text
function Readline:GetRegister(register) end

---@param text string
function Readline:Insert(text) end

function Readline:Log() end

function Readline:Prompt() end

---@return string keystroke
function Readline:ReadChar() end

---@param register string
---@param text string
function Readline:SetRegister(register, text) end

---@return string? input
function Readline:read() end

function Readline:refreshPrompt() end

---@param fn fun(line:string,pos:integer):table,string
function Readline:setCompleter(fn) end

---@param fn fun(line:string):string
function Readline:setHighlighter(fn) end

---@param fn fun(line:string,pos:integer):string
function Readline:setHinter(fn) end

---@param handler table
function Readline:setHistory(handler) end

---@param mode string
function Readline:setInputMode(mode) end

---@param fn fun(...: any)
function Readline:setRawInputCallback(fn) end

---@param fn fun(needle:string,haystack:table<string>):table|nil
function Readline:setSearcher(fn) end

---@param fn fun(...: any)
function Readline:setViActionCallback(fn) end

---@param fn fun(...: any)
function Readline:setViModeCallback(fn) end

---@class readline
---@field fuzzySearch fun(needle: string, haystack: table): table
---@field new fun(): Readline
---@field newHistory fun(path: string): table
local readline = {}

return readline
