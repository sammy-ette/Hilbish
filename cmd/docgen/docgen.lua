local fs = require 'fs'
local emmyPattern = '^%-%-%- (.+)'
local emmyPattern2 = '^%-%- (.+)'
local modpattern = '^%-+ @module (.+)'
local pieces = {}
local descriptions = {}

local files = fs.readdir 'nature'
for _, fname in ipairs(files) do
	local isScript = fname:match'%.lua$'
	if not isScript then goto continue end

	local f = io.open(string.format('nature/%s', fname))
	local header = f:read '*l'
	local mod = header:match(modpattern)
	if not mod then goto continue end

	pieces[mod] = {}
	descriptions[mod] = {}

	local docPiece = {}
	local lines = {}
	local lineno = 0
	local doingDescription = true

	for line in f:lines() do
		lineno = lineno + 1
		lines[lineno] = line

		if line == header then goto continue2 end
		if line:match(emmyPattern) or line:match(emmyPattern2) then
			if doingDescription then
				table.insert(descriptions[mod], line:match(emmyPattern) or line:match(emmyPattern2))
			end
		else
			doingDescription = false
			if line:match '^function' then
				local pattern = (string.format('^function %s%%.', mod) .. '(%w+)')
				local funcName = line:match(pattern)
				if not funcName then goto continue2 end

				local dps = {
					description = {},
					example = {},
					params = {}
				}

				local offset = 1
				local doingExample = false
				while true do
					local prev = lines[lineno - offset]

					local docline = prev:match '^%-+ (.+)'
					if docline then
						local emmy = docline:match '@(%w+)'
						local cut = 0

						if emmy then cut = emmy:len() + 3 end
						local emmythings = string.split(docline:sub(cut), ' ')

						if emmy then
							if emmy == 'param' then
								table.insert(dps.params, 1, {
									name = emmythings[1],
									type = emmythings[2],
									-- the +1 accounts for space.
									description = table.concat(emmythings, ' '):sub(emmythings[1]:len() + 1 + emmythings[2]:len() + 1)
								})
							end
						else
							if docline:match '#example' then
								doingExample = not doingExample
							end

							if not docline:match '#example' then
								if doingExample then
										table.insert(dps.example, 1, docline)
								else
									table.insert(dps.description, 1, docline)
								end
							end
						end
						offset = offset + 1
					else
						break
					end
				end

				table.insert(pieces[mod], {funcName, dps})
			end
			docPiece = {}
			goto continue2
		end

		table.insert(docPiece, line)
		::continue2::
	end
	::continue::
end

local header = [[---
title: %s %s
description: %s
layout: doc
menu:
  docs:
    parent: "%s"
---

]]

for iface, dps in pairs(pieces) do
	local mod = iface ~= 'nature' and iface:match '(%w+)' or 'nature'
	local docParent = 'Nature'

	path = string.format('docs/%s/%s.md', mod, iface)
	if mod ~= 'nature' then
		docParent = "API"
		path = string.format('docs/api/%s/%s.md', mod, iface)
	end
	if iface == 'hilbish' then
		docParent = "API"
		path = string.format('docs/api/hilbish/_index.md', mod, iface)
	end

	fs.mkdir(fs.dir(path), true)

	local exists = pcall(fs.stat, path)
	local newOrNotNature = (exists and mod ~= 'nature') or iface == 'hilbish'

	--local f <close> = io.open(path, newOrNotNature and 'r+' or 'w+')
	if not newOrNotNature then
		--f:write(string.format(header, 'Module', iface, (descriptions[iface] and #descriptions[iface] > 0) and descriptions[iface][1] or 'No description.', docParent))
		if descriptions[iface] and #descriptions[iface] > 0 then
			table.remove(descriptions[iface], 1)
			--f:write(string.format('\n## Introduction\n%s\n\n', table.concat(descriptions[iface], '\n')))
			--f:write('## Functions\n')
		end
	end

	print(mod, dps)
	table.sort(dps, function(a, b) return a[1] < b[1] end)
	--[[for _, piece in pairs(dps) do
		local func = piece[1]
		local docs = piece[2]
		print(func, docs)
	end]]
end
