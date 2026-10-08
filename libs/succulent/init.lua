-- string functions
string.split = function(str, delimiter)
	local res = {}
	local from = 1

	local delimFrom, delimTo = string.find(str, delimiter, from)

	while delimFrom do
		table.insert(res, string.sub(str, from, delimFrom - 1))
		from = delimTo + 1
		delimFrom, delimTo = string.find(str, delimiter, from)
	end

	table.insert(res, string.sub(str, from))

	return res
end

-- table functions
table.filter = function(t, fn)
	local res = {}

	for k, v in pairs(t) do
		if fn(v, k, t) then table.insert(res, v) end
	end

	return res
end

table.map = function(t, fn)
	local res = {}

	for k, v in pairs(t) do
		table.insert(res, fn(v, k, t))
	end

	return res
end

