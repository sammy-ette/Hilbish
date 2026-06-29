local inspect = require 'inspect'
local M = {}

M.lastExit = 0
M.lastErr = nil

function M.exec(cmdstr)
    -- go through snail directly (not hilbish.run) so we keep res.err
    local res = hilbish.snail:run(cmdstr, {})
    M.lastExit, M.lastErr = res.exitCode, res.err
    return res.exitCode
end

function M.capture(cmdstr)
    local code, out = hilbish.run(cmdstr, false)
    M.lastExit = code
    return (out:gsub('%s+$', ''))
end

function M.result(cmdstr)
    local exitCode, stdout, stderr = hilbish.run(cmdstr, false)
    M.lastExit = exitCode
    return {
        exitCode = exitCode,
        stdout = stdout,
        stderr = stderr
    }
end

function M.env(var)
    return env[var] or ''
end

function M.dispatch(cmd, args, cmdString)
    local luaEnvVar = _ENV[cmd]
    if luaEnvVar and type(luaEnvVar) == 'function' then
        M.lastExit = 0
        return luaEnvVar(table.unpack(args))
    else
        return M.exec(cmdString)
    end
end

function M.quote(value)
    if type(value) == 'table' then
        local parts = {}
        for _, v in ipairs(value) do parts[#parts + 1] = M.quote(v) end
        return table.concat(parts, ' ')
    end
    return "'" .. tostring(value):gsub("'", "'\\''") .. "'"
end

function M.show(v)
    if v == nil then return end
    if type(v) == 'string' then print(v) else print(inspect(v)) end
end

function M.thru(acc, ...)
    for i = 1, select('#', ...) do
        local rhs = select(i, ...)
        if type(rhs) == 'function' then
            acc = rhs(acc)
        else
            acc = acc | rhs
        end
    end
    return acc
end

function M.feed(cmdstr, input)
    local insink = hilbish.sink.new()
    insink:write(input == nil and '' or tostring(input))
    local outsink = hilbish.sink.new()
    local res = hilbish.snail:run(cmdstr, { sinks = { input = insink, out = outsink } })
    M.lastExit, M.lastErr = res.exitCode, res.err
    return (outsink:readAll():gsub('%s+$', ''))
end

function M.shellfn(cmdstr)
    return function(input) return M.feed(cmdstr, input) end
end

-- pipefn resolves a bare-name pipe stage at runtime: a callable global is a Lua
-- function (functional pipe), `nil` means it's a shell command (input as stdin),
-- anything else is a bitwise-or
function M.pipefn(name)
    return function(input)
        local v = _ENV[name]
        if type(v) == 'function' then return v(input) end
        if v == nil then return M.feed(name, input) end
        return input | v
    end
end

function M.pipeline(stages)
    local input, exit = nil, 0
    local luaRet
    local n = #stages
    for idx, stage in ipairs(stages) do
        local last = idx == n
        if stage.lua then
            local prevStage = stages[idx - 1]
            if prevStage and prevStage.lua then
                input = luaRet
            end
            local out = stage.lua(input or '')
            luaRet = out
            out = out ~= nil and tostring(out) or ''
            if last then
                M.show(luaRet)
            else
                input = out
            end
        else
            local sinks = {}
            if input ~= nil then
                local insink = hilbish.sink.new()
                insink:write(input)
                sinks.input = insink
            end
            local outsink
            if not last then
                outsink = hilbish.sink.new()
                sinks.out = outsink
            end
            local res = hilbish.snail:run(stage.shell, { sinks = sinks })
            exit = (res and res.exitCode) or 0
            if not last then input = outsink:readAll() end
        end
    end
    M.lastExit = exit
    return exit
end

return M
