package luaruntime

import (
	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v3"
)

// yamlEncode encodes a Lua value to a YAML string.
// Lua usage: str = yaml.encode(data)
func (rt *LuaRuntime) yamlEncode(L *lua.LState) int {
	lv := L.CheckAny(1)
	goValue := luaValueToGoValue(lv)

	yamlData, err := yaml.Marshal(goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(yamlData)))
	return 1
}

// yamlDecode decodes a YAML string into a Lua table.
// Lua usage: data = yaml.decode(str)
func (rt *LuaRuntime) yamlDecode(L *lua.LState) int {
	yamlStr := L.CheckString(1)

	var goValue interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &goValue); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(goValueToLuaValue(L, goValue))
	return 1
}
