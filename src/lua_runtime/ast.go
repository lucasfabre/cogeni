package luaruntime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucasfabre/cogeni/src/astparser"
	lua "github.com/yuin/gopher-lua"
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
// @param language string|nil The grammar name (optional, defaults to file extension).
// @returns table The AST table.
// </lua_api>
func (rt *LuaRuntime) cogeniReadAST(L *lua.LState) int {
	firstArg := L.CheckAny(1)
	lang := L.OptString(2, "")

	var source []byte
	var err error

	if firstArg.Type() == lua.LTString {
		// Case 2: It's a path string
		path := firstArg.String()

		// If lang is not provided, try to detect it from extension
		if lang == "" {
			ext := filepath.Ext(path)
			lang = rt.cfg.GetGrammarForExtension(ext)
			if lang == "" {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("could not detect language for extension '%s'", ext)))
				return 2
			}
		}
		currentFile := rt.L.GetGlobal("_CURRENT_FILE").String()

		// Resolve path relative to current file if it is relative
		if !filepath.IsAbs(path) && currentFile != "nil" && currentFile != "" {
			path = filepath.Join(filepath.Dir(currentFile), path)
		}

		// Validate path to prevent traversal
		var err error
		path, err = rt.validatePath(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if rt.WaitFunc != nil {
			if err := rt.WaitFunc(path, currentFile); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("dependency error for '%s': %v", path, err)))
				return 2
			}
		}

		// Try reading from buffer first
		if rt.ReadFunc != nil {
			if content, ok := rt.ReadFunc(path); ok {
				source = []byte(content)
			}
		}

		if source == nil {
			source, err = os.ReadFile(path)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(fmt.Sprintf("failed to read file '%s': %v", firstArg.String(), err)))
				return 2
			}
		}
	} else {
		// Case 1: Assume it's an object with a 'read' method
		readFunc := L.GetField(firstArg, "read")
		if readFunc.Type() != lua.LTFunction {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("invalid first argument: expected string or object with 'read' method, got %s", firstArg.Type().String())))
			return 2
		}

		L.Push(readFunc)
		L.Push(firstArg)
		L.Push(lua.LString("*a"))
		L.Call(2, 1)

		result := L.Get(-1)
		L.Pop(1)

		if result == lua.LNil {
			L.Push(lua.LNil)
			L.Push(lua.LString("failed to read from handle (returned nil)"))
			return 2
		}
		source = []byte(result.String())
	}

	ast, err := rt.parser.Parse(lang, source)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("parsing error: %v", err)))
		return 2
	}

	// Convert astparser.Node to Lua table
	luaAST := rt.nodeToLuaTable(ast)
	L.Push(luaAST)
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
func (rt *LuaRuntime) cogeniRegisterGrammar(L *lua.LState) int {
	name := L.CheckString(1)
	url := L.CheckString(2)
	optsTable := L.ToTable(3)

	opts := astparser.GrammarOptions{}
	if optsTable != nil {
		opts.BuildCmd = L.GetField(optsTable, "build_cmd").String()
		opts.Artifact = L.GetField(optsTable, "artifact").String()
		opts.Branch = L.GetField(optsTable, "branch").String()

		// gopher-lua .String() returns "nil" if the field is absent and we didn't check type
		if L.GetField(optsTable, "build_cmd").Type() == lua.LTNil {
			opts.BuildCmd = ""
		}
		if L.GetField(optsTable, "artifact").Type() == lua.LTNil {
			opts.Artifact = ""
		}
		if L.GetField(optsTable, "branch").Type() == lua.LTNil {
			opts.Branch = ""
		}
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
func (rt *LuaRuntime) cogeniGetGrammar(L *lua.LState) int {
	ext := L.CheckString(1)
	grammar := rt.cfg.GetGrammarForExtension(ext)
	L.Push(lua.LString(grammar))
	return 1
}

// nodeToLuaTable recursively converts an astparser.Node to a Lua table.
// This table is what Lua scripts interact with when querying the AST.
func (rt *LuaRuntime) nodeToLuaTable(node astparser.Node) *lua.LTable {
	tbl := rt.L.CreateTable(0, 0)

	tbl.RawSetString("id", lua.LNumber(node.Id))
	tbl.RawSetString("type", lua.LString(node.Type))
	tbl.RawSetString("kind_id", lua.LNumber(node.KindId))
	tbl.RawSetString("grammar_id", lua.LNumber(node.GrammarId))
	tbl.RawSetString("grammar_name", lua.LString(node.GrammarName))
	if node.Content != "" {
		tbl.RawSetString("content", lua.LString(node.Content))
	}
	tbl.RawSetString("is_named", lua.LBool(node.IsNamed))
	tbl.RawSetString("is_extra", lua.LBool(node.IsExtra))
	tbl.RawSetString("has_error", lua.LBool(node.HasError))
	tbl.RawSetString("is_error", lua.LBool(node.IsError))
	tbl.RawSetString("is_missing", lua.LBool(node.IsMissing))

	startPos := rt.L.CreateTable(2, 0)
	startPos.Append(lua.LNumber(node.StartPos[0]))
	startPos.Append(lua.LNumber(node.StartPos[1]))
	tbl.RawSetString("start_pos", startPos)

	endPos := rt.L.CreateTable(2, 0)
	endPos.Append(lua.LNumber(node.EndPos[0]))
	endPos.Append(lua.LNumber(node.EndPos[1]))
	tbl.RawSetString("end_pos", endPos)

	if len(node.Children) > 0 {
		childrenTable := rt.L.CreateTable(len(node.Children), 0)
		for _, child := range node.Children {
			childrenTable.Append(rt.nodeToLuaTable(child))
		}
		tbl.RawSetString("children", childrenTable)
	}

	if len(node.Fields) > 0 {
		fieldsTable := rt.L.CreateTable(0, len(node.Fields))
		for key, val := range node.Fields {
			fieldsTable.RawSetString(key, rt.convertASTValueToLua(val))
		}
		tbl.RawSetString("fields", fieldsTable)
	}

	return tbl
}

func (rt *LuaRuntime) convertASTValueToLua(val any) lua.LValue {
	switch v := val.(type) {
	case astparser.Node:
		return rt.nodeToLuaTable(v)
	case []astparser.Node:
		listTable := rt.L.CreateTable(len(v), 0)
		for _, item := range v {
			listTable.Append(rt.nodeToLuaTable(item))
		}
		return listTable
	case string:
		return lua.LString(v)
	case float64:
		return lua.LNumber(v)
	case bool:
		return lua.LBool(v)
	default:
		return lua.LString(fmt.Sprintf("%v", v))
	}
}
