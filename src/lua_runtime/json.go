package luaruntime

import (
	"encoding/json"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// jsonEncode encodes a Lua value to a JSON string.
// It supports an optional table for formatting options.
// Lua usage: str = json.encode(data, { indent = true })
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

// goValueToLuaValue converts a Go interface{} to a lua.LValue.
func goValueToLuaValue(L *lua.LState, val interface{}) lua.LValue {
	switch v := val.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}:
		tbl := L.CreateTable(len(v), 0)
		for _, item := range v {
			tbl.Append(goValueToLuaValue(L, item))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.CreateTable(0, len(v))
		for k, val := range v {
			tbl.RawSetString(k, goValueToLuaValue(L, val))
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", v))
	}
}

// luaValueToGoValue converts a lua.LValue to a Go interface{}.
func luaValueToGoValue(lv lua.LValue) interface{} {
	if lv == lua.LNil {
		return nil
	}

	switch v := lv.(type) {
	case *lua.LTable:
		// Check if it's an array or a map
		if v.Len() > 0 && v.RawGetInt(1) != lua.LNil {
			arr := make([]interface{}, 0, v.Len())
			for i := 1; i <= v.Len(); i++ {
				arr = append(arr, luaValueToGoValue(v.RawGetInt(i)))
			}
			return arr
		} else {
			m := make(map[string]interface{})
			v.ForEach(func(key, val lua.LValue) {
				m[key.String()] = luaValueToGoValue(val)
			})
			return m
		}
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	case *lua.LFunction, *lua.LUserData, *lua.LState, *lua.LChannel:
		return fmt.Sprintf("unsupported_lua_type:%T", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
