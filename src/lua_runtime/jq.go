package luaruntime

import (
	"fmt"

	"github.com/itchyny/gojq"
	lua "github.com/yuin/gopher-lua"
)

// jqQuery executes a JQ query against a Lua value (typically an AST table).
// It is exposed to Lua as jq.query(data, query_string).
//
// data: Any Lua value (table, string, number, etc.) to be queried.
// query_string: A standard JQ query string (e.g., ".children[] | select(.type == 'class')").
//
// Returns: A single Lua value if the query has one result, or a table of values
// if there are multiple results. Returns nil if no matches are found.
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
