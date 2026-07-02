-- @module hilbish.aliases
-- command aliasing
-- The alias interface manages command aliases in Hilbish. An alias is a short name
-- that expands to a longer command when entered at the prompt.
-- Aliases are stored as-is in history (the short name, not the expanded command).
--
-- Aliases support numbered argument substitution. `%1` in the alias body is replaced
-- by the first argument, `%2` by the second, and so on. `%0` is passed through
-- literally. A substitution can be escaped by prefixing it with `\`.
---@diagnostic disable-next-line: missing-fields
hilbish.aliases = {
    all = {}
}

--- This is an alias (ha) for the [hilbish.alias](../#alias) function.
--- @param alias string
--- @param cmd string
--- @since 2.0
--- @see hilbish.alias
function hilbish.aliases.add(alias, cmd)
    hilbish.aliases.all[alias] = cmd
end

--- Removes an alias.
--- @param alias string
function hilbish.aliases.delete(alias)
    hilbish.aliases.all[alias] = nil
end

--- Returns a table of all defined aliases.
--- Keys are the alias names and values are the command strings they expand to.
--- @return table<string, string> aliases A table mapping alias names to their command strings.
--- @example
--- hilbish.aliases.add('hi', 'echo hi')
--- local aliases = hilbish.aliases.list()
--- -- -> {hi = 'echo hi'}
--- @example
function hilbish.aliases.list()
    return hilbish.aliases.all
end

--- Resolves an alias to its original command. Will thrown an error if the alias doesn't exist.
--- @param cmdstr string
--- @return string resolved
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
