package luaruntime

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
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
// @param options table Optional formatting options (e.g. `{indent=true}`).
// @returns string The JSON string.
// </lua_api>
func (rt *LuaRuntime) jsonEncode(L *lua.LState) int {
	lv := L.CheckAny(1)
	goValue := luaValueToGoValue(lv)

	indent := ""
	if L.GetTop() >= 2 {
		options := L.CheckTable(2)
		if options.RawGetString("indent") == lua.LTrue {
			indent = "  "
		}
	}

	var jsonData []byte
	var err error

	if indent != "" {
		jsonData, err = json.MarshalIndent(goValue, "", indent)
	} else {
		jsonData, err = json.Marshal(goValue)
	}

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(jsonData)))
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
func (rt *LuaRuntime) jsonDecode(L *lua.LState) int {
	jsonStr := L.CheckString(1)

	var goValue interface{}
	if err := json.Unmarshal([]byte(jsonStr), &goValue); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(goValueToLuaValue(L, goValue))
	return 1
}
