package luaruntime

import (
	"os"
	"path/filepath"

	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
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
func (rt *LuaRuntime) fsFind(L *luajit.State) int {
	dir := L.CheckString(1)

	fileType := ""
	namePattern := ""
	maxDepth := -1

	// Check options table
	if L.GetTop() >= 2 && !L.IsNil(2) {
		L.PushValue(2) // Push table to top

		L.GetField(-1, "type")
		if L.IsString(-1) {
			fileType = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(-1, "name")
		if L.IsString(-1) {
			namePattern = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(-1, "maxdepth")
		if L.IsNumber(-1) {
			maxDepth = int(L.ToNumber(-1))
		}
		L.Pop(1)

		L.Pop(1) // Pop table copy
	}

	L.CreateTable(0, 0) // Result table
	tableIdx := L.GetTop()

	baseDir, _ := filepath.Abs(dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		absPath, _ := filepath.Abs(path)
		rel, _ := filepath.Rel(baseDir, absPath)
		depth := 0
		if rel != "." {
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

		if info.IsDir() && info.Name() == "framework" {
			return filepath.SkipDir
		}

		if fileType == "f" && info.IsDir() {
			return nil
		}
		if fileType == "d" && !info.IsDir() {
			return nil
		}

		if namePattern != "" && namePattern != "nil" {
			matched, err := filepath.Match(namePattern, info.Name())
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		L.PushBool(true)
		L.SetField(tableIdx, path)
		return nil
	})

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}

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
func (rt *LuaRuntime) fsBaseDir(L *luajit.State) int {
	path := L.CheckString(1)
	L.PushString(filepath.Dir(path))
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
func (rt *LuaRuntime) fsBaseName(L *luajit.State) int {
	path := L.CheckString(1)
	L.PushString(filepath.Base(path))
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
func (rt *LuaRuntime) fsPathJoin(L *luajit.State) int {
	top := L.GetTop()
	parts := make([]string, top)
	for i := 1; i <= top; i++ {
		parts[i-1] = L.CheckString(i)
	}
	L.PushString(filepath.Join(parts...))
	return 1
}
