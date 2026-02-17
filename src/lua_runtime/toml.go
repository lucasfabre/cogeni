package luaruntime

import (
	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
	"github.com/pelletier/go-toml/v2"
)

// tomlEncode encodes a Lua value to a TOML string.
// Lua usage: str = toml.encode(data)
//
// <lua_api>
// @module toml
// @function encode
// @summary Encodes a Lua value to a TOML string.
// @usage toml.encode(val)
// @param val any The value to encode.
// @returns string The TOML string.
// </lua_api>
func (rt *LuaRuntime) tomlEncode(L *luajit.State) int {
	goValue := ToGoValue(L, 1)

	tomlData, err := toml.Marshal(goValue)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	L.PushString(string(tomlData))
	return 1
}

// tomlDecode decodes a TOML string into a Lua table.
// Lua usage: data = toml.decode('foo = 1')
//
// <lua_api>
// @module toml
// @function decode
// @summary Decodes a TOML string into a Lua table.
// @usage toml.decode(str)
// @param str string The TOML string.
// @returns any The decoded Lua value.
// </lua_api>
func (rt *LuaRuntime) tomlDecode(L *luajit.State) int {
	tomlStr := L.CheckString(1)

	var goValue interface{}
	if err := toml.Unmarshal([]byte(tomlStr), &goValue); err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	PushGoValue(L, goValue)
	return 1
}
