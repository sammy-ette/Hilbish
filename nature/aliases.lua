-- @module hilbish.aliases
-- command aliasing
-- The alias interface deals with all command aliases in Hilbish.
hilbish.aliases = {
    all = {}
}

--- This is an alias (ha) for the [hilbish.alias](../#alias) function.
--- @param alias string
--- @param cmd string
function hilbish.aliases.add(alias, cmd)
    hilbish.aliases.all[alias] = cmd
end

--- Removes an alias.
--- @param alias string
function hilbish.aliases.delete(alias)
    hilbish.aliases.all[alias] = nil
end

-- Get a table of all aliases, with string keys as the alias and the value as the command.
--[[
    #example
    hilbish.aliases.add('hi', 'echo hi')
    
    local aliases = hilbish.aliases.list()
    -- -> {hi = 'echo hi'}
    #example
--]]
--- @return table[string, string]
function hilbish.aliases.list()
    return hilbish.aliases.all
end

--- Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.
--- @param cmdstr string
--- @return string
function hilbish.aliases.resolve(cmdstr)
    local args = string.split(cmdstr, ' ')
    if #args == 0 then
        -- again, this shouldnt be possible. im just porting it from go
        return cmdstr
    end

    local visited = {}
    while hilbish.aliases.all[args[1]] ~= nil do
        if visited[args[1]] then
            break
        end
        visited[args[1]] = true

        local alias = hilbish.aliases.all[args[1]]
        alias = alias:gsub('\\?%%%d+', function(match)
            local idx = tonumber(match:match('%d+'))
            if match:sub(1, 1) == '\\' or idx == 0 then
                -- unescape or skip %0
                return (match:gsub('^\\', ''))
            end

            if idx + 1 > #args then
                return match
            end

            local val = args[idx+1]
            table.remove(args, idx+1)
            cmdstr = table.concat(args, ' ')
            return val
        end)

        cmdstr = alias .. cmdstr:sub(#args[1] + 1)
        args = string.split(cmdstr, ' ')
    end

    return cmdstr
end
