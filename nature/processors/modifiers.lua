hilbish.processors.add {
    priority = -9999,
    name = 'modifiers',
    ---@param input string
    func = function(input)
        local modifiers = {}

        while true do
            local mod, rest = input:match '^(@%S+)%s*(.*)'
            if not mod then
                break
            end

            local name = mod:match '^@([a-zA-Z0-9]+)'
            local valRaw = mod:match '^@[a-zA-Z0-9]+=(.+)'
            local val
            if not valRaw then
                val = true
            else
                val = valRaw

                if valRaw == 'false' then val = false end
                if valRaw == 'true' then val = true end
            end

            modifiers[name] = val

            input = rest
        end

        return {
            command = input,
            modifiers = modifiers
        }
    end
}
