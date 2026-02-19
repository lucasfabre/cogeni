package luaruntime

import (
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// fsFind recursively searches for files and directories matching specific criteria.
// It is exposed to Lua as fs.find(dir, options).
//
// <lua_api>
// @module fs
// @function find
// @summary Recursively finds files and directories.
// @usage fs.find(dir, options)
// @param dir string The starting directory.
// @param options table Filters (type="f"|"d", name="glob", maxdepth=N).
// @returns table A table where keys are paths and values are 'true'.
// </lua_api>
func (rt *LuaRuntime) fsFind(L *lua.LState) int {
	dir := L.CheckString(1)
	options := L.OptTable(2, L.CreateTable(0, 0))

	fileType := options.RawGetString("type").String()
	namePattern := options.RawGetString("name").String()
	maxDepthVal := options.RawGetString("maxdepth")
	maxDepth := -1
	if maxDepthVal.Type() == lua.LTNumber {
		maxDepth = int(maxDepthVal.(lua.LNumber))
	}

	files := rt.L.CreateTable(0, 0)

	absDir, err := rt.validatePath(dir)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	baseDir := absDir

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		absPath, _ := filepath.Abs(path)
		rel, _ := filepath.Rel(baseDir, absPath)
		depth := 0
		if rel != "." {
			// Count separators to determine depth
			depth = 1
			for _, char := range rel {
				if char == os.PathSeparator {
					depth++
				}
			}
		}

		if maxDepth >= 0 && depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip framework directory
		if info.IsDir() && info.Name() == "framework" {
			return filepath.SkipDir
		}

		// Filter by type
		if fileType == "f" && info.IsDir() {
			return nil
		}
		if fileType == "d" && !info.IsDir() {
			return nil
		}

		// Filter by name pattern (glob)
		if namePattern != "" && namePattern != "nil" {
			matched, err := filepath.Match(namePattern, info.Name())
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		files.RawSetString(path, lua.LTrue)
		return nil
	})

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(files)
	return 1
}

// fsBaseDir returns the directory of a path.
// Lua usage: dir = fs.basedir("a/b/c.txt") -> "a/b"
//
// <lua_api>
// @module fs
// @function basedir
// @summary Returns the directory of a path.
// @usage fs.basedir(path)
// @param path string The file path.
// @returns string The directory part of the path.
// </lua_api>
func (rt *LuaRuntime) fsBaseDir(L *lua.LState) int {
	path := L.CheckString(1)
	L.Push(lua.LString(filepath.Dir(path)))
	return 1
}

// fsBaseName returns the last element of a path.
// Lua usage: name = fs.basename("a/b/c.txt") -> "c.txt"
//
// <lua_api>
// @module fs
// @function basename
// @summary Returns the last element of a path.
// @usage fs.basename(path)
// @param path string The file path.
// @returns string The file name.
// </lua_api>
func (rt *LuaRuntime) fsBaseName(L *lua.LState) int {
	path := L.CheckString(1)
	L.Push(lua.LString(filepath.Base(path)))
	return 1
}

// fsPathJoin joins multiple path elements.
// Lua usage: path = fs.join("a", "b", "c") -> "a/b/c"
//
// <lua_api>
// @module fs
// @function join
// @summary Joins multiple path elements.
// @usage fs.join(...)
// @param ... string Path elements to join.
// @returns string The joined path.
// </lua_api>
func (rt *LuaRuntime) fsPathJoin(L *lua.LState) int {
	top := L.GetTop()
	parts := make([]string, top)
	for i := 1; i <= top; i++ {
		parts[i-1] = L.CheckString(i)
	}
	L.Push(lua.LString(filepath.Join(parts...)))
	return 1
}
