package luaruntime

import (
	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
	"gopkg.in/yaml.v3"
)

// yamlEncode encodes a Lua value to a YAML string.
// Lua usage: str = yaml.encode(data)
//
// <lua_api>
// @module yaml
// @function encode
// @summary Encodes a Lua value to a YAML string.
// @usage yaml.encode(val)
// @param val any The value to encode.
// @returns string The YAML string.
// </lua_api>
func (rt *LuaRuntime) yamlEncode(L *luajit.State) int {
	goValue := ToGoValue(L, 1)

	yamlData, err := yaml.Marshal(goValue)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	L.PushString(string(yamlData))
	return 1
}

// yamlDecode decodes a YAML string into a Lua table.
// Lua usage: data = yaml.decode('foo: 1')
//
// <lua_api>
// @module yaml
// @function decode
// @summary Decodes a YAML string into a Lua table.
// @usage yaml.decode(str)
// @param str string The YAML string.
// @returns any The decoded Lua value.
// </lua_api>
func (rt *LuaRuntime) yamlDecode(L *luajit.State) int {
	yamlStr := L.CheckString(1)

	var goValue interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &goValue); err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	PushGoValue(L, goValue)
	return 1
}
