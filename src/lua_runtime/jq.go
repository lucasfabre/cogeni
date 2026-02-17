package luaruntime

import (
	"github.com/itchyny/gojq"
	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
)

// jqQuery runs a JQ query against a Lua value.
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
func (rt *LuaRuntime) jqQuery(L *luajit.State) int {
	goSource := ToGoValue(L, 1)
	queryStr := L.CheckString(2)

	query, err := gojq.Parse(queryStr)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	iter := query.Run(goSource)

	var results []interface{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			L.PushNil()
			L.PushString(err.Error())
			return 2
		}
		results = append(results, v)
	}

	if len(results) == 0 {
		L.PushNil()
		return 1
	}

	if len(results) == 1 {
		PushGoValue(L, results[0])
		return 1
	}

	L.CreateTable(len(results), 0)
	tableIdx := L.GetTop()
	for i, res := range results {
		PushGoValue(L, res)
		L.RawSetInt(tableIdx, i+1)
	}

	return 1
}
