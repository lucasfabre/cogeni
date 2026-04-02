local T = require("tests/lua/framework/test_helper")
local ROOT = os.getenv("COGENI_TEST_ROOT") or io.popen("pwd"):read("*l")

T.describe("jq integration", function()
	T.it("should support jq", function()
		local json_data = [[
        {
            "a": 1,
            "b": 2,
            "c": {
                "d": 3
            },
            "e": [
                1,
                2,
                3
            ]
        }
        ]]
		local obj = json.decode(json_data)
		T.assert(obj ~= nil, "failed to decode json")

		local d = jq.query(obj, ".c.d")
		T.assert_eq(d, 3, "failed to query jq")

		local e = jq.query(obj, ".e")
		T.assert_eq(e, { 1, 2, 3 }, "failed to query jq list")
	end)

	T.it("should query asts", function()
		local ast = cogeni.read_ast(fs.join(ROOT, "examples/fullstack/frontend/types.ts"), "typescript")
		T.assert(ast ~= nil, "failed to read ast")

		local query = [[
          [.children[] | select(.type == "export_statement") | .fields.declaration | select(.type == "interface_declaration")]
          | map(.fields.name.content)
        ]]
		local names = jq.query(ast, query)
		T.assert_eq(names, { "User", "Product", "Todo" }, "failed to query interface names")
	end)
end)
