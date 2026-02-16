---
sidebar_position: 1
---

# Lua API Reference

The `cogeni` module provides the core functionality for reading ASTs and controlling the generation process. It also exposes several utility modules for file system operations, JSON manipulation, and JQ querying.

## Core Module (`cogeni`)

The `cogeni` module is the entry point for most operations.

### `read_ast(path)`

Parses a source file using the configured Tree-sitter grammar and returns its Abstract Syntax Tree (AST) as a Lua table.

**Arguments:**
- `path` (string): The path to the source file to parse.

**Returns:**
- `table`: A Lua table representing the AST. Nodes have `type`, `start_byte`, `end_byte`, and `children`.

**Example:**
```lua
local ast = cogeni.read_ast("src/main.py")
print(ast.root.type) -- "module"
```

### `process(path)`

Recursively triggers processing for another file. This is crucial for building dependency chains and ensuring that required files are processed before dependent ones.

**Arguments:**
- `path` (string): The path to the file to process.

**Example:**
```lua
cogeni.process("src/models.py")
-- Code below runs after src/models.py has been processed if there's a dependency
```

### `outfile(id, path)`

Registers a file as an output target for a specific ID. Any content written to this ID using `write()` will be saved to this file.

**Arguments:**
- `id` (string): A unique identifier for this output stream.
- `path` (string): The destination file path.

**Example:**
```lua
cogeni.outfile("models", "generated/models.py")
write("models", "class User: pass")
```

### `outtag(id, path, tag)`

Registers a specific tagged block within an existing file as an output target. This allows `cogeni` to inject code into files without overwriting the entire file.

**Arguments:**
- `id` (string): A unique identifier for this output stream.
- `path` (string): The path to the file containing the tag.
- `tag` (string): The name of the tag block (e.g., `<cogeni:mytag>`).

**Example:**
```lua
-- In src/hooks.py:
-- # <cogeni:hooks>
-- # </cogeni:hooks>

cogeni.outtag("hooks", "src/hooks.py", "hooks")
write("hooks", "def pre_save(): pass")
```

## File System Module (`fs`)

The `fs` module provides utilities for file system traversal and path manipulation.

### `find(pattern)`

Finds files matching a glob pattern.

**Arguments:**
- `pattern` (string): The glob pattern (e.g., `src/**/*.py`).

**Returns:**
- `table`: A list of file paths.

**Example:**
```lua
local files = fs.find("src/**/*.py")
for _, file in ipairs(files) do
    print(file)
end
```

### `join(part1, part2, ...)`

Joins path components into a single path string, handling platform-specific separators.

**Arguments:**
- `part...` (string): Path components.

**Returns:**
- `string`: The joined path.

**Example:**
```lua
local path = fs.join("src", "models", "user.py")
```

### `basedir(path)`

Returns the directory component of a path.

**Arguments:**
- `path` (string): The file path.

**Returns:**
- `string`: The directory path.

**Example:**
```lua
local dir = fs.basedir("src/models/user.py") -- "src/models"
```

### `basename(path)`

Returns the filename component of a path.

**Arguments:**
- `path` (string): The file path.

**Returns:**
- `string`: The filename.

**Example:**
```lua
local name = fs.basename("src/models/user.py") -- "user.py"
```

## JSON Module (`json`)

The `json` module provides encoding and decoding of JSON data.

### `encode(table)`

Encodes a Lua table into a JSON string.

**Arguments:**
- `table` (table): The Lua table to encode.

**Returns:**
- `string`: The JSON string.

**Example:**
```lua
local data = { name = "cogeni", version = 1 }
local json_str = json.encode(data)
```

### `decode(string)`

Decodes a JSON string into a Lua table.

**Arguments:**
- `string` (string): The JSON string to decode.

**Returns:**
- `table`: The decoded Lua table.

**Example:**
```lua
local data = json.decode('{"name": "cogeni"}')
print(data.name)
```

## JQ Module (`jq`)

The `jq` module allows you to run JQ queries directly on Lua tables or JSON strings. This is powerful for extracting data from ASTs or configuration files.

### `query(input, filter)`

Executes a JQ filter on the input.

**Arguments:**
- `input` (table or string): The input data (Lua table or JSON string).
- `filter` (string): The JQ filter string.

**Returns:**
- `table` or `string` or `number` or `boolean`: The result of the query.

**Example:**
```lua
local ast = cogeni.read_ast("config.json")
local names = jq.query(ast, ".users[].name")
```

## Global Functions

### `write(id, content)`

Buffers generated content for a specific target ID. The ID must be registered via `outfile` or `outtag` first.

**Arguments:**
- `id` (string): The target identifier.
- `content` (string): The content to write.

**Example:**
```lua
write("output_id", "content to write\n")
```
