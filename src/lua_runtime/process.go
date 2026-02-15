package luaruntime

import (
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// cogeniProcess triggers the generation lifecycle for a specific file.
// It is exposed to Lua as cogeni.process(path).
//
// path: The filesystem path of the file to process. If relative, it is resolved
// against the directory of the current script.
//
// This function allows for recursive and multi-file orchestration. It invokes
// the coordinator's processing logic, which may run the target file's own
// <cogeni> blocks if they exist.
//
// Returns: true on success, or (nil, error_string) on failure.
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
