---
sidebar_position: 2
---

# Lua API

`cogeni` exposes several global modules to the Lua environment, allowing you to interact with the file system, parse ASTs, and generate code.

## Core Module (`cogeni`)

The `cogeni` module provides the core functionality for reading ASTs and controlling the generation process.

### `read_ast(path)`
Parses a file and returns its AST as a Lua table.

```lua
local ast = cogeni.read_ast("src/main.py")
```

### `process(path)`
Recursively triggers processing for another file. This is useful for building dependency chains.

```lua
cogeni.process("src/models.py")
```

### `outfile(id, path)`
Directs `write(id, content)` output to a whole file.

```lua
cogeni.outfile("models", "generated/models.py")
write("models", "class User: pass")
```

### `outtag(id, path, tag)`
Directs `write(id, content)` output to a tagged block in a file. This allows for modifying existing files.

```lua
cogeni.outtag("hooks", "src/hooks.py", "pre_save")
write("hooks", "def pre_save(): pass")
```

## File System (`fs`)

Utilities for file system operations.

- `find(pattern)`: Finds files matching a pattern.
- `join(part1, part2, ...)`: Joins path components.
- `basedir(path)`: Returns the directory of a path.
- `basename(path)`: Returns the filename of a path.

## JSON (`json`)

JSON encoding and decoding utilities.

- `encode(table)`: Encodes a Lua table to a JSON string.
- `decode(string)`: Decodes a JSON string to a Lua table.

## JQ (`jq`)

Execute JQ queries on Lua tables or JSON strings.

```lua
local result = jq.query(data, ".users[] | .name")
```

## Global Functions

### `write(id, content)`
Buffers generated content for a specific target ID (defined by `outfile` or `outtag`).

```lua
write("output_id", "content to write")
```
