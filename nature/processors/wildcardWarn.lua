local hilbish = require 'hilbish'
local Greenhouse = require 'nature.greenhouse'
local Page = require 'nature.greenhouse.page'

print 'wildcard warn loaded'

local function contains(search, needle)
	for _, p in ipairs(search) do
		if p:match(needle) then
			return p
		end
	end

	return nil
end

local stdoutSink = {}
function stdoutSink:write(t)
    io.write(t)
end

function stdoutSink:writeln(t)
    print(t)
end

hilbish.processors.add {
    name = 'wildcardWarn',
    func = function(commandLine)
        local args = string.split(commandLine, ' ')
        if args[1] == 'rm' then -- command check
            local match1 = contains(args, '%*')
            local match2 = contains(args, '/%*')
            if match1 or match2 then
                local matches = fs.glob(match1 or match2)
                if #matches == 0 then
                    return {continue = true}
                end

                ::askWithInfo::
                print 'Detected wildcard with potentially dangerous command.'
                print 'Are you sure you want to run this command?'

                ::ask::
                local ans = hilbish.read '(Y/N, L to list files matched with wildcard) '
                ans = ans:lower()
                if ans == 'l' then
                    local gh = Greenhouse(stdoutSink)
                    local page = Page('Wildcard File List', '')
                    page.lines = matches

                    gh:addPage(page)
                    gh:initUi()
                    goto askWithInfo
                elseif ans == 'y' then
                    return {continue = true}
                elseif ans == 'n' then
                    return {continue = false}
                else
                    print 'Invalid answer..'
                    goto ask
                end
            end
        end
    end
}