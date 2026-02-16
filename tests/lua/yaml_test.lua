local T = require("tests/lua/framework/test_helper")

T.describe("yaml.encode", function()
	T.it("should encode a table to yaml", function()
		local t = { a = 1, b = "hello", c = { true, false } }
		local s = yaml.encode(t)
		T.assert(s:find("a: 1"), "missing key a")
		T.assert(s:find("b: hello"), "missing key b")
	end)
end)

T.describe("yaml.decode", function()
	T.it("should decode yaml to a table", function()
		local s = [[
a: 1
b: hello
c:
  - true
  - false
]]
		local t = yaml.decode(s)
		T.assert_eq(t.a, 1, "key a mismatch")
		T.assert_eq(t.b, "hello", "key b mismatch")
		T.assert_eq(t.c[1], true, "key c[1] mismatch")
		T.assert_eq(t.c[2], false, "key c[2] mismatch")
	end)
end)
