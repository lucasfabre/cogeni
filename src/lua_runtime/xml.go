package luaruntime

import (
	"github.com/clbanning/mxj/v2"
	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
)

// xmlEncode encodes a Lua value to an XML string.
// Lua usage: str = xml.encode(data, { root = "myroot" })
//
// <lua_api>
// @module xml
// @function encode
// @summary Encodes a Lua value to an XML string.
// @usage xml.encode(val, options)
// @param val any The value to encode.
// @param options table Optional formatting options (e.g. {root="root"}).
// @returns string The XML string.
// </lua_api>
func (rt *LuaRuntime) xmlEncode(L *luajit.State) int {
	goValue := ToGoValue(L, 1)

	rootName := "root"
	if L.GetTop() >= 2 && L.IsTable(2) {
		L.GetField(2, "root")
		if L.IsString(-1) {
			rootName = L.ToString(-1)
		}
		L.Pop(1)
	}

	m, ok := goValue.(map[string]interface{})
	if !ok {
		m = map[string]interface{}{
			"#text": goValue,
		}
	}

	wrapper := map[string]interface{}{
		rootName: m,
	}

	mv := mxj.Map(wrapper)
	xmlData, err := mv.XmlIndent("", "  ")
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	L.PushString(string(xmlData))
	return 1
}

// xmlDecode decodes an XML string into a Lua table.
// Lua usage: data = xml.decode('<root>foo</root>')
//
// <lua_api>
// @module xml
// @function decode
// @summary Decodes an XML string into a Lua table.
// @usage xml.decode(str)
// @param str string The XML string.
// @returns any The decoded Lua value.
// </lua_api>
func (rt *LuaRuntime) xmlDecode(L *luajit.State) int {
	xmlStr := L.CheckString(1)

	mv, err := mxj.NewMapXml([]byte(xmlStr))
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	PushGoValue(L, map[string]interface{}(mv))
	return 1
}
