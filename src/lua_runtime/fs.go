package luaruntime

import (
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// fsFind recursively searches for files and directories matching specific criteria.
// It is exposed to Lua as fs.find(dir, options).
//
// dir: The starting directory for the search.
// options: A table containing optional filters:
//   - type: "f" for files, "d" for directories.
//   - name: Glob pattern for matching filenames (e.g., "*.lua").
//   - maxdepth: Maximum recursion depth (integer).
//
// Returns: A table where keys are the relative paths found and values are 'true'.
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

	baseDir, _ := filepath.Abs(dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
func (rt *LuaRuntime) fsBaseDir(L *lua.LState) int {
	path := L.CheckString(1)
	L.Push(lua.LString(filepath.Dir(path)))
	return 1
}

// fsBaseName returns the last element of a path.
// Lua usage: name = fs.basename("a/b/c.txt") -> "c.txt"
func (rt *LuaRuntime) fsBaseName(L *lua.LState) int {
	path := L.CheckString(1)
	L.Push(lua.LString(filepath.Base(path)))
	return 1
}

// fsPathJoin joins multiple path elements.
// Lua usage: path = fs.join("a", "b", "c") -> "a/b/c"
func (rt *LuaRuntime) fsPathJoin(L *lua.LState) int {
	top := L.GetTop()
	parts := make([]string, top)
	for i := 1; i <= top; i++ {
		parts[i-1] = L.CheckString(i)
	}
	L.Push(lua.LString(filepath.Join(parts...)))
	return 1
}
