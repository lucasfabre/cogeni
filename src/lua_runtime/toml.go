package luaruntime

import (
	"github.com/pelletier/go-toml/v2"
	lua "github.com/yuin/gopher-lua"
)

// tomlEncode encodes a Lua value to a TOML string.
// Lua usage: str = toml.encode(data)
func (rt *LuaRuntime) tomlEncode(L *lua.LState) int {
	lv := L.CheckAny(1)
	goValue := luaValueToGoValue(lv)

	tomlData, err := toml.Marshal(goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(tomlData)))
	return 1
}

// tomlDecode decodes a TOML string into a Lua table.
// Lua usage: data = toml.decode(str)
func (rt *LuaRuntime) tomlDecode(L *lua.LState) int {
	tomlStr := L.CheckString(1)

	var goValue interface{}
	if err := toml.Unmarshal([]byte(tomlStr), &goValue); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(goValueToLuaValue(L, goValue))
	return 1
}
