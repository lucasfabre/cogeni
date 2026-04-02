local T = require("tests/lua/framework/test_helper")
local ROOT = os.getenv("COGENI_TEST_ROOT") or io.popen("pwd"):read("*l")

T.describe("cogeni.read_ast", function()
	T.it("should read from a file handle", function()
		local f = io.open(fs.join(ROOT, "examples/fullstack/frontend/api.ts"), "r")
		T.assert(f ~= nil, "failed to open test file")
		local ast, err = cogeni.read_ast(f, "typescript")
		f:close()
		T.assert(err == nil, "error reading ast: " .. (err or ""))
		T.assert(ast ~= nil, "ast is nil")
		T.assert_eq(ast.type, "program")
	end)

	T.it("should read from a file path", function()
		local ast, err = cogeni.read_ast(fs.join(ROOT, "examples/fullstack/frontend/api.ts"), "typescript")
		T.assert(err == nil, "error reading ast: " .. (err or ""))
		T.assert(ast ~= nil, "ast is nil")
		T.assert_eq(ast.type, "program")
	end)
end)

T.describe("json.encode", function()
	T.it("should encode a table", function()
		local t = { a = 1, b = "hello", c = { true, false } }
		local s = json.encode(t)
		T.assert(s:find('"a":%s*1'), "missing key a")
		T.assert(s:find('"b":%s*"hello"'), "missing key b")
	end)

	T.it("should support indentation", function()
		local t = { a = 1 }
		local s = json.encode(t, { indent = true })
		T.assert(s:find("\n  "), "should contain indentation")
	end)

	T.it("should register grammars", function()
		cogeni.register_grammar("sql", "https://github.com/DerekStride/tree-sitter-sql", {
			branch = "gh-pages",
			build_cmd = "cc -shared -fPIC -I./src src/parser.c src/scanner.c -o sql.so",
			artifact = "sql.so",
		})
		local ast, err = cogeni.read_ast(fs.join(ROOT, "examples/fullstack/db/schema.sql"), "sql")
		T.assert(err == nil, "error reading ast: " .. (err or ""))
		T.assert(ast ~= nil, "ast is nil")
		T.assert_eq(ast.type, "program")
	end)
end)

T.describe("fs path utilities", function()
	T.it("should get basedir", function()
		T.assert_eq(T.normalize_path(fs.basedir("a/b/c.txt")), "a/b")
		T.assert_eq(fs.basedir("a.txt"), ".")
	end)

	T.it("should get basename", function()
		T.assert_eq(fs.basename("a/b/c.txt"), "c.txt")
	end)

	T.it("should join paths", function()
		T.assert_eq(T.normalize_path(fs.join("a", "b", "c.txt")), "a/b/c.txt")
	end)
end)
