package luaruntime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucasfabre/codegen/src/astparser"
	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
)

// cogeniReadAST parses source code into a detailed AST.
// It is exposed to Lua as cogeni.read_ast(source, language).
//
// <lua_api>
// @module cogeni
// @function read_ast
// @summary Parses source code into a detailed AST.
// @usage cogeni.read_ast(source, language)
// @param source string|table File path or handle.
// @param language string The grammar name (e.g. "go", "python").
// @returns table The AST table.
// </lua_api>
func (rt *LuaRuntime) cogeniReadAST(L *luajit.State) int {
	firstArg := ToGoValue(L, 1)
	lang := L.CheckString(2)

	var source []byte
	var err error

	if path, ok := firstArg.(string); ok {
		// Case 2: It's a path string
		L.GetGlobal("_CURRENT_FILE")
		currentFile := ""
		if L.IsString(-1) {
			currentFile = L.ToString(-1)
		}
		L.Pop(1)

		if !filepath.IsAbs(path) && currentFile != "" && currentFile != "nil" {
			path = filepath.Join(filepath.Dir(currentFile), path)
		}

		if rt.WaitFunc != nil {
			if err := rt.WaitFunc(path, currentFile); err != nil {
				L.PushNil()
				L.PushString(fmt.Sprintf("dependency error for '%s': %v", path, err))
				return 2
			}
		}

		if rt.ReadFunc != nil {
			if content, ok := rt.ReadFunc(path); ok {
				source = []byte(content)
			}
		}

		if source == nil {
			source, err = os.ReadFile(path)
			if err != nil {
				L.PushNil()
				L.PushString(fmt.Sprintf("failed to read file '%s': %v", firstArg.(string), err))
				return 2
			}
		}
	} else {
		// Case 1: Assume it's an object with a 'read' method
		if !L.IsTable(1) && !L.IsUserData(1) { // UserData might also have metatable
			// Just check getfield
		}

		L.GetField(1, "read")
		if !L.IsFunction(-1) {
			L.PushNil()
			L.PushString(fmt.Sprintf("invalid first argument: expected string or object with 'read' method"))
			return 2
		}

		L.PushValue(1)     // self
		L.PushString("*a") // arg
		L.Call(2, 1)       // Call read(self, "*a") -> 1 result

		// Result is at top
		if L.IsNil(-1) {
			L.PushNil()
			L.PushString("failed to read from handle (returned nil)")
			return 2
		}
		source = []byte(L.ToString(-1))
		L.Pop(1) // Pop result
	}

	ast, err := rt.parser.Parse(lang, source)
	if err != nil {
		L.PushNil()
		L.PushString(fmt.Sprintf("parsing error: %v", err))
		return 2
	}

	rt.nodeToLuaTable(L, ast)
	return 1
}

// cogeniRegisterGrammar registers a custom grammar source URL.
// Lua usage: cogeni.register_grammar("name", "url", { branch = "...", build_cmd = "..." })
//
// <lua_api>
// @module cogeni
// @function register_grammar
// @summary Registers a custom grammar source URL.
// @usage cogeni.register_grammar(name, url, opts)
// @param name string The grammar name.
// @param url string The git repository URL.
// @param opts table Options (branch, build_cmd, artifact).
// </lua_api>
func (rt *LuaRuntime) cogeniRegisterGrammar(L *luajit.State) int {
	name := L.CheckString(1)
	url := L.CheckString(2)

	opts := astparser.GrammarOptions{}
	if L.GetTop() >= 3 && L.IsTable(3) {
		L.GetField(3, "build_cmd")
		if L.IsString(-1) {
			opts.BuildCmd = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(3, "artifact")
		if L.IsString(-1) {
			opts.Artifact = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(3, "branch")
		if L.IsString(-1) {
			opts.Branch = L.ToString(-1)
		}
		L.Pop(1)
	}

	rt.parser.RegisterGrammar(name, url, opts)
	return 0
}

// cogeniGetGrammar looks up a grammar name by file extension.
// Lua usage: grammar = cogeni.get_grammar(".py")
//
// <lua_api>
// @module cogeni
// @function get_grammar
// @summary Looks up a grammar name by file extension.
// @usage cogeni.get_grammar(ext)
// @param ext string The file extension (including dot).
// @returns string The grammar name.
// </lua_api>
func (rt *LuaRuntime) cogeniGetGrammar(L *luajit.State) int {
	ext := L.CheckString(1)
	grammar := rt.cfg.GetGrammarForExtension(ext)
	L.PushString(grammar)
	return 1
}

func (rt *LuaRuntime) nodeToLuaTable(L *luajit.State, node astparser.Node) {
	L.CreateTable(0, 16) // Pre-allocate
	tableIdx := L.GetTop()

	L.PushNumber(float64(node.Id))
	L.SetField(tableIdx, "id")

	L.PushString(node.Type)
	L.SetField(tableIdx, "type")

	L.PushNumber(float64(node.KindId))
	L.SetField(tableIdx, "kind_id")

	L.PushNumber(float64(node.GrammarId))
	L.SetField(tableIdx, "grammar_id")

	L.PushString(node.GrammarName)
	L.SetField(tableIdx, "grammar_name")

	if node.Content != "" {
		L.PushString(node.Content)
		L.SetField(tableIdx, "content")
	}

	L.PushBool(node.IsNamed)
	L.SetField(tableIdx, "is_named")

	L.PushBool(node.IsExtra)
	L.SetField(tableIdx, "is_extra")

	L.PushBool(node.HasError)
	L.SetField(tableIdx, "has_error")

	L.PushBool(node.IsError)
	L.SetField(tableIdx, "is_error")

	L.PushBool(node.IsMissing)
	L.SetField(tableIdx, "is_missing")

	L.CreateTable(2, 0)
	L.PushNumber(float64(node.StartPos[0]))
	L.RawSetInt(-2, 1)
	L.PushNumber(float64(node.StartPos[1]))
	L.RawSetInt(-2, 2)
	L.SetField(tableIdx, "start_pos")

	L.CreateTable(2, 0)
	L.PushNumber(float64(node.EndPos[0]))
	L.RawSetInt(-2, 1)
	L.PushNumber(float64(node.EndPos[1]))
	L.RawSetInt(-2, 2)
	L.SetField(tableIdx, "end_pos")

	if len(node.Children) > 0 {
		L.CreateTable(len(node.Children), 0)
		childrenIdx := L.GetTop()
		for i, child := range node.Children {
			rt.nodeToLuaTable(L, child)
			L.RawSetInt(childrenIdx, i+1)
		}
		L.SetField(tableIdx, "children")
	}

	if len(node.Fields) > 0 {
		L.CreateTable(0, len(node.Fields))
		fieldsIdx := L.GetTop()
		for key, val := range node.Fields {
			rt.convertASTValueToLua(L, val)
			L.SetField(fieldsIdx, key)
		}
		L.SetField(tableIdx, "fields")
	}
}

func (rt *LuaRuntime) convertASTValueToLua(L *luajit.State, val any) {
	switch v := val.(type) {
	case astparser.Node:
		rt.nodeToLuaTable(L, v)
	case []astparser.Node:
		L.CreateTable(len(v), 0)
		listIdx := L.GetTop()
		for i, item := range v {
			rt.nodeToLuaTable(L, item)
			L.RawSetInt(listIdx, i+1)
		}
	case string:
		L.PushString(v)
	case float64:
		L.PushNumber(v)
	case bool:
		L.PushBool(v)
	default:
		L.PushString(fmt.Sprintf("%v", v))
	}
}
