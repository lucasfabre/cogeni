package luaruntime

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// goValueToLuaValue converts a Go interface{} to a lua.LValue.
func goValueToLuaValue(L *lua.LState, val interface{}) lua.LValue {
	switch v := val.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case float32:
		return lua.LNumber(float64(v))
	case int:
		return lua.LNumber(float64(v))
	case int32:
		return lua.LNumber(float64(v))
	case int64:
		return lua.LNumber(float64(v))
	case uint:
		return lua.LNumber(float64(v))
	case uint32:
		return lua.LNumber(float64(v))
	case uint64:
		return lua.LNumber(float64(v))
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
	case map[interface{}]interface{}:
		tbl := L.CreateTable(0, len(v))
		for k, val := range v {
			tbl.RawSet(goValueToLuaValue(L, k), goValueToLuaValue(L, val))
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
