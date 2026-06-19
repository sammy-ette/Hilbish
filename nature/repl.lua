local bait = require 'bait'
local terminal = require 'terminal'

while hilbish.interactive do
	::rerun::
	hilbish.running = false

	local ok, res = pcall(function() return hilbish.editor:read() end)
	if not ok and tostring(res):lower():match 'eof' then
		bait.throw 'hilbish.exit'
		os.exit(0)
	end
	if not ok then
		if tostring(res):lower():match 'ctrl%+c' then
			print '^C'
			bait.throw 'hilbish.cancel'
		else
			print(tostring(res))
			_ = io.read()
		end
		goto continue
	end
	---@type string|nil
	local input = res

	local priv = false
	if res:sub(1, 1) == ' ' then
		priv = true
	end
	---@diagnostic disable-next-line: need-check-nil
	input = input:gsub('%s+$', '')
	--:gsub('^([%s]+).', '')

	if input:len() == 0 then
		hilbish.running = true
		bait.throw('command.exit', 0 )
		goto continue
	end

	if input:match '\\$' then
		io.write '\n'
		while true do
			input = hilbish.runner.continuePrompt(input:gsub('\\$', '') .. '\n', false)
			if not input then
				goto rerun
			end

			if not input:match '\\$' then break end
		end
	end

	hilbish.running = true
	hilbish.runner.run(input, priv)

	local ok, term = pcall(function() return terminal.size() end)
	if ok and term and term.width and term.width > 0 then
		io.write(string.char(0x001b) .. '[7m∆' .. string.char(0x001b) .. '[0m' .. string.rep(' ', term.width - 1) .. "\r")
		io.flush()
	end

	::continue::
end
