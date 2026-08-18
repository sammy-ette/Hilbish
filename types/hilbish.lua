---@meta

---@class hilbish.abbr
---@field add fun(abbr: string, expanded: string|fun(...: any), opts?: table)
---@field remove fun(abbr: string)

---@class hilbish.aliases
---@field add fun(alias: string, cmd: string)
---@field delete fun(alias: string)
---@field list fun(): table<string, string>
---@field resolve fun(cmdstr: string): string

---@class hilbish.completions
---@field add fun(scope: string, cb: fun(query:string,ctx:string,fields:table<string>):table,string)
---@field bins fun(query: string, ctx: string, fields: table): table<string>, string
---@field call fun(name: string, query: string, ctx: string, fields: table): table, string
---@field dirs fun(query: string, ctx: string, fields: table): table<string>, string
---@field files fun(query: string, ctx: string, fields: table): table<string>, string
---@field handler fun(line: string, pos: number): string, table

---@class Job
---@field cmd any
---@field running any
---@field suspended any
---@field id any
---@field pid any
---@field exitCode any
---@field stdout any
---@field stderr any
local Job = {}

function Job:background() end

function Job:foreground() end

---@param opts? table
---@return number exitCode
function Job:start(opts) end

function Job:stop() end

---@class hilbish.jobs
---@field add fun(cmdstr: string, opts: table): Job
---@field all fun(): table<Job>
---@field disown fun(id: number)
---@field get fun(id: number): Job?
---@field last fun(): Job?
---@field stopAll fun()

---@class hilbish.message
---@field icon any
---@field title any
---@field text any
---@field channel any
---@field summary any
---@field index any
---@field read any
hilbish.message = {}

---@class hilbish.messageIcons
---@field INFO any
---@field SUCCESS any
---@field WARN any
---@field ERROR any
hilbish.messageIcons = {}

---@class hilbish.messages
---@field all fun(): table
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
---@field add fun(processor: table)
---@field execute fun(command: string, opts?: table): table

---@class Runner
---@field run any
---@field validate any
local Runner = {}

---@class RunnerResult
---@field exitCode any
---@field input any
---@field err any
---@field continue any
---@field newline any
local RunnerResult = {}

---@class hilbish.runner
---@field add fun(name: string, runner: Runner)
---@field exec fun(cmd: string, runnerName?: string): RunnerResult
---@field get fun(name: string): Runner
---@field getCurrent fun(): string
---@field lua fun(input: string): RunnerResult
---@field run fun(input: string, priv?: boolean)
---@field set fun(name: string, runner: Runner)
---@field setCurrent fun(name: string)
---@field sh fun(input: string): RunnerResult

---@class Sink
local Sink = {}

---@param auto? boolean
function Sink:autoFlush(auto) end

function Sink:flush() end

---@return string line
function Sink:read() end

---@return string data
function Sink:readAll() end

---@param str string
function Sink:write(str) end

---@param str string
function Sink:writeln(str) end

---@class hilbish.sink

---@class Timer
---@field type any
---@field running any
---@field duration any
---@field id any
local Timer = {}

function Timer:start() end

function Timer:stop() end

---@class hilbish.timers
---@field INTERVAL any
---@field TIMEOUT any
---@field create fun(type: number, time: number, callback: fun(...: any)): Timer
---@field get fun(id: number): Timer?
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
---@field midnightEdition any
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
---@field sink hilbish.sink
---@field timers hilbish.timers
---@field userDir hilbish.userDir
---@field alias fun(alias: string, cmd: string)
---@field appendPath fun(path: string|table)
---@field cwd fun(): string
---@field exec fun(cmd: string)
---@field interval fun(cb: fun(...: any), time: number): Timer
---@field lookpath fun(file: string): string
---@field multiprompt fun(str?: string): string?
---@field prependPath fun(path: string|table)
---@field prompt fun(p: string, typ?: string)
---@field read fun(prompt?: string): string?
---@field run fun(cmd: string, streams: table|boolean): number, string?, string?
---@field timeout fun(cb: fun(...: any), time: number): Timer
---@field which fun(name: string): string?

---@type Hilbish
---@diagnostic disable-next-line: missing-fields
hilbish = {}

return hilbish
