local T = require("tests/lua/framework/test_helper")

T.describe("xml.encode", function()
	T.it("should encode a table to xml", function()
		local t = { a = 1, b = "hello" }
		local s = xml.encode(t, { root = "root" })
		T.assert(s:find("<root>"), "missing root tag")
		T.assert(s:find("<a>1</a>"), "missing element a")
		T.assert(s:find("<b>hello</b>"), "missing element b")
	end)

	T.it("should encode attributes", function()
		local t = { ["-id"] = "123", ["#text"] = "content" }
		local s = xml.encode(t, { root = "item" })
		T.assert(s:find('<item id="123">content</item>'), "incorrect attribute encoding")
	end)

	T.it("should encode nested elements", function()
		local t = { child = { sub = "val" } }
		local s = xml.encode(t, { root = "parent" })
		T.assert(s:find("<parent>"), "missing parent")
		T.assert(s:find("<child>"), "missing child")
		T.assert(s:find("<sub>val</sub>"), "missing sub")
	end)
end)

T.describe("xml.decode", function()
	T.it("should decode xml to a table", function()
		local s = [[<root><a>1</a><b>hello</b></root>]]
		local t = xml.decode(s)
		-- print("DEBUG: " .. json.encode(t))
		T.assert_eq(t.root.a, "1", "key a mismatch")
		T.assert_eq(t.root.b, "hello", "key b mismatch")
	end)

	T.it("should decode attributes and text", function()
		local s = [[<item id="123">content</item>]]
		local t = xml.decode(s)
		T.assert_eq(t.item["-id"], "123", "attribute mismatch")
		T.assert_eq(t.item["#text"], "content", "text mismatch")
	end)

	T.it("should decode lists", function()
		local s = [[<root><item>1</item><item>2</item></root>]]
		local t = xml.decode(s)
		T.assert_eq(t.root.item[1], "1", "item 1 mismatch")
		T.assert_eq(t.root.item[2], "2", "item 2 mismatch")
	end)
end)
