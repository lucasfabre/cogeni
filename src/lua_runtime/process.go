package luaruntime

import (
	"path/filepath"

	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
)

// cogeniProcess triggers the generation lifecycle for a specific file.
// Lua usage: cogeni.process("path/to/file.py")
//
// <lua_api>
// @module cogeni
// @function process
// @summary Triggers the generation lifecycle for a specific file.
// @usage cogeni.process(path)
// @param path string The file path to process.
// @returns boolean True if successful.
// </lua_api>
func (rt *LuaRuntime) cogeniProcess(L *luajit.State) int {
	path := L.CheckString(1)
	if rt.ProcessFunc == nil {
		L.PushNil()
		L.PushString("cogeni.process is not available in this context")
		return 2
	}

	L.GetGlobal("_CURRENT_FILE")
	currentFile := ""
	if L.IsString(-1) {
		currentFile = L.ToString(-1)
	}
	L.Pop(1)

	// Resolve path relative to current file if it is relative
	if !filepath.IsAbs(path) && currentFile != "nil" && currentFile != "" {
		path = filepath.Join(filepath.Dir(currentFile), path)
	}

	if err := rt.ProcessFunc(path, currentFile); err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

	L.PushBool(true)
	return 1
}
