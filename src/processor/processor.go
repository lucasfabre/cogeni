package processor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/clbanning/mxj/v2"
	"github.com/lucasfabre/codegen/src/config"
	luaruntime "github.com/lucasfabre/codegen/src/lua_runtime"
	lua "github.com/yuin/gopher-lua"
)

// ProcessFile executes the end-to-end generation lifecycle for a single file.
func ProcessFile(cfg *config.Config, filePath string, c *Coordinator) error {
	content, err := c.readFileContent(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	luaScript := ExtractCogeniBlocks(string(content))
	absPath, _ := filepath.Abs(filePath)
	if luaScript == "" {
		if c.CleanMode {
			return c.ApplyChangesToBuffer(absPath, nil, false)
		}
		return nil // No blocks found to execute.
	}

	// Acquire a runtime from the pool (or create new if pool is empty)
	rt, err := c.runtimePool.Acquire()
	if err != nil {
		return fmt.Errorf("failed to initialize Lua runtime: %w", err)
	}
	defer c.runtimePool.Release(rt)

	rt.ProcessFunc = func(path, requestor string) error {
		return c.Process(path, requestor)
	}
	rt.WaitFunc = func(path, requestor string) error {
		return c.WaitForReader(path, requestor)
	}
	rt.ReadFunc = c.GetResult

	rt.L.SetGlobal("_CURRENT_FILE", lua.LString(absPath))
	rt.L.SetGlobal("_FILE_EXTENSION", lua.LString(filepath.Ext(filePath)))

	if err := rt.DoString(luaScript); err != nil {
		return fmt.Errorf("script execution failed for '%s': %w", filePath, err)
	}

	rt.Schedule()
	return c.CaptureResults(rt, absPath)
}

// CaptureResults processes the output collected by a LuaRuntime and maps it to target files.
func (c *Coordinator) CaptureResults(rt *luaruntime.LuaRuntime, sourcePath string) error {
	fileChanges := make(map[string]map[string]string)
	fullFileOverwrites := make(map[string]bool)

	for id, out := range rt.Output {
		targetPath := sourcePath
		effectiveTag := id
		isFullFile := false

		if target, ok := rt.OutputTargets[id]; ok {
			baseDir := "."
			if sourcePath != "" {
				baseDir = filepath.Dir(sourcePath)
			}
			resolvedPath := target.Path
			if !filepath.IsAbs(resolvedPath) {
				resolvedPath = filepath.Join(baseDir, resolvedPath)
			}
			absTarget, err := filepath.Abs(resolvedPath)
			if err == nil {
				targetPath = absTarget
			}

			if target.Tag == "" {
				isFullFile = true
			} else {
				effectiveTag = target.Tag
			}
		}

		if targetPath == "" {
			continue
		}

		if _, ok := fileChanges[targetPath]; !ok {
			fileChanges[targetPath] = make(map[string]string)
		}
		fileChanges[targetPath][effectiveTag] = out
		if isFullFile {
			fullFileOverwrites[targetPath] = true
		}

		if sourcePath != "" && targetPath != sourcePath {
			c.mu.RLock()
			t, ok := c.tasks[sourcePath]
			c.mu.RUnlock()
			if ok {
				t.addDependency(targetPath)
			}
		}
	}

	for path, outputs := range fileChanges {
		isFull := fullFileOverwrites[path]
		if err := c.ApplyChangesToBuffer(path, outputs, isFull); err != nil {
			return err
		}
	}

	return nil
}

// ApplyChangesToBuffer merges new outputs into the existing content of a file.
func (c *Coordinator) ApplyChangesToBuffer(path string, outputs map[string]string, isFullFile bool) error {
	var currentContent string

	c.mu.RLock()
	buffered, ok := c.Results[path]
	c.mu.RUnlock()

	if ok {
		currentContent = buffered
	} else {
		content, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) && !isFullFile {
				return fmt.Errorf("failed to read target file '%s': %w", path, err)
			}
			currentContent = ""
		} else {
			currentContent = string(content)
		}
	}

	var newContent string
	if isFullFile {
		if c.CleanMode {
			newContent = ""
		} else {
			for _, out := range outputs {
				newContent = out
				break
			}
		}
	} else {
		effectiveOutputs := outputs
		if c.CleanMode {
			effectiveOutputs = make(map[string]string)
			for id := range outputs {
				effectiveOutputs[id] = ""
			}
		}
		newContent = ReplaceGeneratedBlocks(currentContent, effectiveOutputs, c.CleanMode)
	}

	c.mu.Lock()
	c.Results[path] = newContent
	c.mu.Unlock()
	return nil
}

// ExtractCogeniBlocks finds all Lua scripts embedded within <cogeni>...</cogeni> tags.
func ExtractCogeniBlocks(content string) string {
	var script strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))
	inside := false
	var prefix string
	var indent string

	for scanner.Scan() {
		line := scanner.Text()

		if !inside {
			if strings.Contains(line, "<cogeni>") && !strings.Contains(line, "<generated") {
				inside = true
				idx := strings.Index(line, "<cogeni>")
				prefix = line[:idx]

				// Calculate indent (leading whitespace of prefix)
				indent = ""
				for _, r := range prefix {
					if unicode.IsSpace(r) {
						indent += string(r)
					} else {
						break
					}
				}
			}
			continue
		}

		if strings.Contains(line, "</cogeni>") {
			inside = false
			continue
		}

		code := line
		// Try to strip the full prefix first (comment char included)
		if strings.HasPrefix(line, prefix) {
			code = line[len(prefix):]
		} else if strings.HasPrefix(line, indent) {
			// Fallback: strip just the indentation
			code = line[len(indent):]
		}
		script.WriteString(code)
		script.WriteString("\n")
	}

	return script.String()
}

// reGenerated is the regex to find <generated ...> tag and capture content inside.
var reGenerated = regexp.MustCompile(`<generated(\s+[^>]*)>`)

// ReplaceGeneratedBlocks replaces the text inside <generated by="cogeni" id="..."> blocks.
func ReplaceGeneratedBlocks(content string, outputs map[string]string, cleanAll bool) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))
	inside := false

	for scanner.Scan() {
		line := scanner.Text()

		if !inside {
			// Check if line contains <generated ...>
			matches := reGenerated.FindStringSubmatch(line)
			if len(matches) > 0 {
				attrStr := matches[1]

				// Parse attributes using mxj by constructing a dummy XML tag
				dummyXML := fmt.Sprintf("<dummy %s />", attrStr)
				mv, err := mxj.NewMapXml([]byte(dummyXML))
				if err == nil {
					// Check attributes
					root, ok := mv["dummy"].(map[string]interface{})
					if ok {
						by, _ := root["-by"].(string)
						id, _ := root["-id"].(string)

						if by == "cogeni" && id != "" {
							result.WriteString(line)
							result.WriteString("\n")

							if replacement, ok := outputs[id]; ok {
								inside = true
								result.WriteString(replacement)
								if !strings.HasSuffix(replacement, "\n") && replacement != "" {
									result.WriteString("\n")
								}
							} else if cleanAll {
								inside = true
							}
							continue
						}
					}
				}
			}

			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		if strings.Contains(line, "</cogeni>") || strings.Contains(line, "</generated>") {
			inside = false
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}
