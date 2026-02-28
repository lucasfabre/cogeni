package luaruntime

import (
	"github.com/clbanning/mxj/v2"
	lua "github.com/yuin/gopher-lua"
)

// xmlEncode encodes a Lua value to an XML string.
// Lua usage: str = xml.encode(data, { root = "root_name" })
//
// <lua_api>
// @module xml
// @function encode
// @summary Encodes a Lua value to an XML string.
// @usage xml.encode(data, options)
// @param data any The Lua value to encode.
// @param options table|nil Optional formatting options (e.g. `{root="root_name"}`).
// @returns string The XML string.
// </lua_api>
func (rt *LuaRuntime) xmlEncode(L *lua.LState) int {
	lv := L.CheckAny(1)
	goValue := luaValueToGoValue(lv)

	rootName := "root"
	if L.GetTop() >= 2 {
		options := L.CheckTable(2)
		if v := options.RawGetString("root"); v != lua.LNil {
			rootName = v.String()
		}
	}

	// Ensure input is a map for mxj
	m, ok := goValue.(map[string]interface{})
	if !ok {
		// If it's not a map (e.g. a list or primitive), wrap it
		m = map[string]interface{}{
			"#text": goValue,
		}
	}

	// Wrap in root element
	wrapper := map[string]interface{}{
		rootName: m,
	}

	mv := mxj.Map(wrapper)
	xmlData, err := mv.XmlIndent("", "  ")
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(xmlData)))
	return 1
}

// xmlDecode decodes an XML string into a Lua table.
// Lua usage: data = xml.decode(str)
//
// <lua_api>
// @module xml
// @function decode
// @summary Decodes an XML string into a Lua table.
// @usage xml.decode(str)
// @param str string The XML string to decode.
// @returns any The decoded Lua value.
// </lua_api>
func (rt *LuaRuntime) xmlDecode(L *lua.LState) int {
	xmlStr := L.CheckString(1)

	mv, err := mxj.NewMapXml([]byte(xmlStr))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(goValueToLuaValue(L, map[string]interface{}(mv)))
	return 1
}
