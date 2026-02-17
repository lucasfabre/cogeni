package luaruntime

import (
	"fmt"

	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
)

// PushGoValue converts a Go interface{} to a Lua value and pushes it to the stack.
func PushGoValue(L *luajit.State, val interface{}) {
	switch v := val.(type) {
	case nil:
		L.PushNil()
	case bool:
		L.PushBool(v)
	case float64:
		L.PushNumber(v)
	case float32:
		L.PushNumber(float64(v))
	case int:
		L.PushNumber(float64(v))
	case int32:
		L.PushNumber(float64(v))
	case int64:
		L.PushNumber(float64(v))
	case uint:
		L.PushNumber(float64(v))
	case uint32:
		L.PushNumber(float64(v))
	case uint64:
		L.PushNumber(float64(v))
	case string:
		L.PushString(v)
	case []interface{}:
		L.CreateTable(len(v), 0)
		for i, item := range v {
			PushGoValue(L, item)
			L.RawSetInt(-2, i+1)
		}
	case map[string]interface{}:
		L.CreateTable(0, len(v))
		for k, val := range v {
			PushGoValue(L, val)
			L.SetField(-2, k)
		}
	case map[interface{}]interface{}:
		L.CreateTable(0, len(v))
		for k, val := range v {
			PushGoValue(L, k)
			PushGoValue(L, val)
			L.SetTable(-3)
		}
	default:
		L.PushString(fmt.Sprintf("%v", v))
	}
}

// ToGoValue converts a Lua value at the given index to a Go interface{}.
func ToGoValue(L *luajit.State, idx int) interface{} {
	absIdx := idx
	if idx < 0 {
		absIdx = L.GetTop() + 1 + idx
	}

	if L.IsNil(absIdx) {
		return nil
	}
	if L.IsBool(absIdx) {
		return L.ToBoolean(absIdx)
	}
	if L.IsNumber(absIdx) {
		return L.ToNumber(absIdx)
	}
	if L.IsString(absIdx) {
		return L.ToString(absIdx)
	}
	if L.IsTable(absIdx) {
		length := L.ObjLen(absIdx)
		if length > 0 {
			L.RawGetInt(absIdx, 1)
			isNil := L.IsNil(-1)
			L.Pop(1)

			if !isNil {
				arr := make([]interface{}, length)
				for i := 0; i < length; i++ {
					L.RawGetInt(absIdx, i+1)
					arr[i] = ToGoValue(L, -1)
					L.Pop(1)
				}
				return arr
			}
		}

		m := make(map[string]interface{})
		L.PushNil()
		for L.Next(absIdx) {
			L.PushValue(-2) // Copy key
			key := L.ToString(-1)
			L.Pop(1)

			val := ToGoValue(L, -1)
			m[key] = val
			L.Pop(1)
		}
		return m
	}

	return nil
}
