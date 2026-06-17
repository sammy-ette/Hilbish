---@meta

---@class Sink
local Sink = {}

function Sink:autoFlush(auto) end

function Sink:flush() end

---@return string
function Sink:read() end

---@return string
function Sink:readAll() end

function Sink:write(str) end

function Sink:writeln(str) end

---@class hilbish.abbr
---@field add fun(abbr: string, expanded: string|fun(...: any), opts: table)
---@field remove fun(abbr: string)

---@class hilbish.aliases
---@field add fun(alias: string, cmd: string)
---@field delete fun(alias: string)
---@field list fun(): table<string, string>
---@field resolve fun(cmdstr: string): string

---@class hilbish.completions
---@field add fun(scope: string, cb: fun(query:string,ctx:string,fields:table<string>):table,string)
---@field bins fun(query: string, ctx: string, fields: table): table, string
---@field call fun(name: string, query: string, ctx: string, fields: table): table, string
---@field dirs fun(query: string, ctx: string, fields: table): table, string
---@field files fun(query: string, ctx: string, fields: table): table, string
---@field handler fun(line: string, pos: number): string, table

---@class Job
---@field cmd any
---@field running any
---@field id any
---@field pid any
---@field exitCode any
---@field stdout any
---@field stderr any
local Job = {}

function Job:background() end

function Job:foreground() end

function Job:start() end

function Job:stop() end

---@class hilbish.jobs
---@field add fun(cmdstr: string, args: table, execPath: string)
---@field all fun(): table<Job>
---@field disown fun(id: number)
---@field get fun(id: any): Job
---@field last fun(): Job
---@field stopAll fun()

---@class hilbish.messages
---@field all fun(): table<hilbish.message>
---@field clear fun()
---@field delete fun(idx: number)
---@field read fun(idx: number)
---@field readAll fun()
---@field send fun(message: hilbish.message)
---@field unreadCount fun(): integer

---@class hilbish.module
---@field paths any
---@field load fun(path: string)

---@class hilbish.os
---@field family any
---@field name any
---@field version any

---@class hilbish.processors
---@field execute fun(command: any, opts: any): table

---@class hilbish.runner
---@field add fun(name: string, runner: table)
---@field exec fun(cmd: string, runnerName: string?): table
---@field get fun(name: string): table
---@field getCurrent fun(): string
---@field lua fun(input: string): table
---@field run fun(input: string, priv: boolean)
---@field set fun(name: string, runner: table)
---@field setCurrent fun(name: string)
---@field sh fun(input: string): table

---@class Timer
---@field type any
---@field running any
---@field duration any
local Timer = {}

function Timer:start() end

function Timer:stop() end

---@class hilbish.timers
---@field INTERVAL any
---@field TIMEOUT any
---@field create fun(type: number, time: number, callback: fun(...: any)): Timer
---@field get fun(id: number): Timer
---@field wait fun()

---@class hilbish.userDir
---@field config any
---@field data any

---@class Hilbish
---@field ver any
---@field goVersion any
---@field user any
---@field host any
---@field dataDir any
---@field defaultConfDir any
---@field confFile any
---@field command any
---@field interactive any
---@field login any
---@field vimMode any
---@field exitCode any
---@field running any
---@field initialized any
---@field home string
---@field editor Readline
---@field snail Snail
---@field history table
---@field opts table
---@field vim table
---@field sink { new: fun(): Sink }
---@field motd string
---@field hinter fun(line: string, pos: number): string
---@field highlighter fun(line: string): string
---@field inputMode fun(mode: string)
---@field appendPath fun(path: string|table)
---@field prependPath fun(path: string|table)
---@field abbr hilbish.abbr
---@field aliases hilbish.aliases
---@field completions hilbish.completions
---@field jobs hilbish.jobs
---@field messages hilbish.messages
---@field module hilbish.module
---@field os hilbish.os
---@field processors hilbish.processors
---@field runner hilbish.runner
---@field timers hilbish.timers
---@field userDir hilbish.userDir
---@field alias fun(alias: string, cmd: string)
---@field cwd fun(): string
---@field exec fun(cmd: string)
---@field interval fun(cb: fun(...: any), time: number): Timer
---@field lookpath fun(file: string): string
---@field multiprompt fun(str: string|nil): string|nil Returns the currently set multilinePrompt if `str` is not provided.
---@field prompt fun(p: string, typ: string)
---@field read fun(prompt: string): string|nil
---@field run fun(cmd: string, streams: table|boolean): number, string, string
---@field timeout fun(cb: fun(...: any), time: number): Timer
---@field which fun(name: string): string|nil

---@type Hilbish
---@diagnostic disable-next-line: missing-fields
hilbish = {}

return hilbish
