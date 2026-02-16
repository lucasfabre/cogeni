local T = require("tests/lua/framework/test_helper")

T.describe("toml.encode", function()
	T.it("should encode a table to toml", function()
		local t = { a = 1, b = "hello" }
		local s = toml.encode(t)
		T.assert(s:find("a = 1.0") or s:find("a = 1"), "missing key a in: " .. s)
		T.assert(s:find('b = "hello"') or s:find("b = 'hello'"), "missing key b in: " .. s)
	end)
end)

T.describe("toml.decode", function()
	T.it("should decode toml to a table", function()
		local s = [[
a = 1
b = "hello"
[c]
d = true
]]
		local t = toml.decode(s)
		T.assert_eq(t.a, 1, "key a mismatch")
		T.assert_eq(t.b, "hello", "key b mismatch")
		T.assert_eq(t.c.d, true, "key c.d mismatch")
	end)
end)
