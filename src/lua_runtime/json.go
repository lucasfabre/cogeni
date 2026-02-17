package luaruntime

import (
	"encoding/json"

	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
)

// jsonEncode encodes a Lua value to a JSON string.
// It supports an optional table for formatting options.
// Lua usage: str = json.encode(data, { indent = true })
//
// <lua_api>
// @module json
// @function encode
// @summary Encodes a Lua value to a JSON string.
// @usage json.encode(val, options)
// @param val any The value to encode.
// @param options table Optional formatting options (e.g. {indent=true}).
// @returns string The JSON string.
// </lua_api>
func (rt *LuaRuntime) jsonEncode(L *luajit.State) int {
	goValue := ToGoValue(L, 1)

	indent := ""
	if L.GetTop() >= 2 && L.IsTable(2) {
		// Check options table for {indent=true}
		L.GetField(2, "indent")
		if L.ToBoolean(-1) {
			indent = "  "
		}
		L.Pop(1)
	}

	var jsonData []byte
	var err error

	if indent != "" {
		jsonData, err = json.MarshalIndent(goValue, "", indent)
	} else {
		jsonData, err = json.Marshal(goValue)
	}

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	L.PushString(string(jsonData))
	return 1
}

// jsonDecode decodes a JSON string into a Lua table.
// Lua usage: data = json.decode('{"foo": 1}')
//
// <lua_api>
// @module json
// @function decode
// @summary Decodes a JSON string into a Lua table.
// @usage json.decode(str)
// @param str string The JSON string.
// @returns any The decoded Lua value.
// </lua_api>
func (rt *LuaRuntime) jsonDecode(L *luajit.State) int {
	jsonStr := L.CheckString(1)

	var goValue interface{}
	if err := json.Unmarshal([]byte(jsonStr), &goValue); err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	PushGoValue(L, goValue)
	return 1
}
