local M = {}

local current_describe = ""
local failures = 0
local passes = 0

function M.describe(name, fn)
	current_describe = name
	print("\n" .. name)
	fn()
end

function M.it(name, fn)
	local start_time = os.clock()
	local ok, err = pcall(fn)
	local end_time = os.clock()
	local elapsed = end_time - start_time

	if ok then
		passes = passes + 1
		print(string.format("  [PASS] %s (%.4fs)", name, elapsed))
	else
		failures = failures + 1
		print(string.format("  [FAIL] %s (%.4fs)", name, elapsed))
		print("    " .. tostring(err))
	end
end

function M.assert(cond, message)
	if not cond then
		error(message or "assertion failed")
	end
end

local function deep_equal(t1, t2)
	if type(t1) ~= type(t2) then
		return false
	end
	if type(t1) ~= "table" then
		return t1 == t2
	end

	for k, v in pairs(t1) do
		if not deep_equal(v, t2[k]) then
			return false
		end
	end
	for k, v in pairs(t2) do
		if t1[k] == nil then
			return false
		end
	end
	return true
end

function M.assert_eq(actual, expected, message)
	if not deep_equal(actual, expected) then
		local actual_str = type(actual) == "table" and json.encode(actual) or tostring(actual)
		local expected_str = type(expected) == "table" and json.encode(expected) or tostring(expected)
		error((message or "") .. "\n    expected: " .. expected_str .. "\n    actual:   " .. actual_str)
	end
end

function M.normalize_path(path)
	return tostring(path):gsub("\\", "/")
end

function M.summary()
	print(string.format("\nTests complete: %d passed, %d failed", passes, failures))
end

function M.exit()
	if failures > 0 then
		os.exit(1)
	else
		os.exit(0)
	end
end

return M
