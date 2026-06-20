local M = {}

function M.exec(cmdstr)
    return hilbish.run(cmdstr, true)
end

function M.capture(cmdstr)
    local _, out = hilbish.run(cmdstr, false)
end

function M.result(cmdstr)
    local exitCode, stdout, stderr = hilbish.run(cmdstr, false)
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
        luaEnvVar(table.unpack(args))
    else
        M.exec(cmdString)
    end
end

return M
