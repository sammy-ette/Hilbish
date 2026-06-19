---@meta

---@class Readline
local Readline = {}

---@param name string
---@param fn fun(...: any)
function Readline:addAction(name, fn) end

---@param key string
---@param action string|fun(...: any)
function Readline:bindKey(key, action) end

---@param amount number
function Readline:deleteByAmount(amount) end

---@return table<string,string>
function Readline:getBindings() end

---@return string
function Readline:getLine() end

---@param register string
---@return string
function Readline:getVimRegister(register) end

---@param text string
function Readline:insert(text) end

function Readline:log(text) end

function Readline:prompt(text) end

---@return string
function Readline:read() end

---@return string
function Readline:readChar() end

---@param name string
function Readline:removeAction(name) end

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

---@param register string
---@param text string
function Readline:setVimRegister(register, text) end

---@param fn fun(needle:string,haystack:table<string>):table|nil
function Readline:setSearcher(fn) end

---@param fn fun(...: any)
function Readline:setViActionCallback(fn) end

---@param fn fun(...: any)
function Readline:setViModeCallback(fn) end

---@param key string
function Readline:unbindKey(key) end

---@class readline
---@field fuzzySearch fun(needle: string, haystack: table): table
---@field new fun(): Readline
---@field newHistory fun(path: string): table
local readline = {}

return readline
