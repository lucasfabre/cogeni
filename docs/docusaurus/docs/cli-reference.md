---
sidebar_position: 2
---

# CLI Reference

`cogeni` provides a set of subcommands to manage the code generation process, from running generators to inspecting dependencies.

## Commands

### `run`

Executes Lua scripts or processes files with embedded `cogeni` blocks. This is the primary command for generation.

**Usage:**
```bash
cogeni run [file...] [flags]
```

**Arguments:**
- `file...`: One or more Lua scripts (`.lua`) or source files with embedded blocks.

**Flags:**
- `--watch`, `-w`: Watch for file changes and re-run.
- `--verbose`, `-v`: Enable verbose logging.

**Example:**
```bash
cogeni run script.lua
cogeni run src/models.py
```

### `clean`

Removes generated content from files, reverting them to their template state. This is useful for resetting the project or testing idempotency.

**Usage:**
```bash
cogeni clean [file...] [flags]
```

**Arguments:**
- `file...`: Files or directories to clean. If omitted, cleans all tracked outputs.

**Flags:**
- `--dry-run`: Show what would be deleted without actually deleting it.

**Example:**
```bash
cogeni clean generated/
```

### `diff`

Shows a diff between the current file content and what `cogeni` would generate. Useful for verifying changes before applying them.

**Usage:**
```bash
cogeni diff [file...] [flags]
```

**Arguments:**
- `file...`: Files to check.

**Example:**
```bash
cogeni diff src/models.py
```

### `ast`

Prints the JSON-serialized AST of the given file. This is invaluable for debugging Lua scripts and understanding the structure of parsed code.

**Usage:**
```bash
cogeni ast <file> [flags]
```

**Arguments:**
- `file`: The source file to parse.

**Example:**
```bash
cogeni ast src/main.py | jq .
```

### `show dependency-graph`

Visualizes the file-level dependencies calculated by the orchestrator. This helps debug circular dependencies or build order issues.

**Usage:**
```bash
cogeni show dependency-graph <file> [flags]
```

**Arguments:**
- `file`: The entry point file.

**Example:**
```bash
cogeni show dependency-graph src/main.py
```

### `repl`

Starts an interactive Lua REPL with all `cogeni` modules pre-loaded.

**Usage:**
```bash
cogeni repl [flags]
```

**Example:**
```lua
> ast = cogeni.read_ast("src/main.py")
> print(ast.root.type)
```

### `config`

Displays or manages the current configuration.

**Usage:**
```bash
cogeni config [flags]
```

**Example:**
```bash
cogeni config
```
