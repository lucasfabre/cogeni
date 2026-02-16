package luaruntime

import (
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// cogeniProcess triggers the generation lifecycle for a specific file.
// It is exposed to Lua as cogeni.process(path).
//
// <lua_api>
// @module cogeni
// @function process
// @summary Triggers the generation lifecycle for a file.
// @usage cogeni.process(path)
// @param path string The file path.
// @returns boolean True on success.
// </lua_api>
func (rt *LuaRuntime) cogeniProcess(L *lua.LState) int {
	path := L.CheckString(1)
	if rt.ProcessFunc == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("cogeni.process is not available in this context"))
		return 2
	}

	currentFile := rt.L.GetGlobal("_CURRENT_FILE").String()

	// Resolve path relative to current file if it is relative
	if !filepath.IsAbs(path) && currentFile != "nil" && currentFile != "" {
		path = filepath.Join(filepath.Dir(currentFile), path)
	}

	if err := rt.ProcessFunc(path, currentFile); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	return 1
}
