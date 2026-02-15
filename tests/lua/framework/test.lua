-- Set up package path to include project root and framework directory
package.path = package.path .. ";./?.lua;./tests/lua/framework/?.lua"

local T = require("tests/lua/framework/test_helper")

-- Load tests
tests = fs.find("./tests/lua", {
	type = "f",
	name = "*.lua",
})

for path in pairs(tests) do
	-- Convert path to module name (e.g., ./tests/lua/api_test.lua -> tests/lua/api_test)
	local module = path:gsub("%.lua$", ""):gsub("^%./", "")
	require(module)
end

-- Output summary
T.summary()
T.exit()
