package luaruntime

import (
	"fmt"

	"github.com/itchyny/gojq"
	lua "github.com/yuin/gopher-lua"
)

// jqQuery executes a JQ query against a Lua value (typically an AST table).
// It is exposed to Lua as jq.query(data, query_string).
//
// <lua_api>
// @module jq
// @function query
// @summary Executes a JQ query against a Lua value.
// @usage jq.query(val, query)
// @param val any The value to query (usually an AST table).
// @param query string The JQ query string.
// @returns any|table A single result or a table of results, or nil.
// </lua_api>
func (rt *LuaRuntime) jqQuery(L *lua.LState) int {
	value := L.CheckAny(1)
	queryString := L.CheckString(2)

	goValue := luaValueToGoValue(value)

	query, err := gojq.Parse(queryString)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("failed to parse jq query: %v", err)))
		return 2
	}

	iter := query.Run(goValue)
	var results []lua.LValue

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("jq query error: %v", err)))
			return 2
		}
		results = append(results, goValueToLuaValue(L, v))
	}

	if len(results) == 0 {
		L.Push(lua.LNil)
		return 1
	}

	if len(results) == 1 {
		L.Push(results[0])
		return 1
	}

	// If multiple results, return as a table
	tbl := L.CreateTable(len(results), 0)
	for _, res := range results {
		tbl.Append(res)
	}
	L.Push(tbl)
	return 1
}
