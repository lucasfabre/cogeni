---
sidebar_position: 8
id: lua-api-reference
title: Lua API Reference
---

# Lua API Reference

The `cogeni` environment provides a suite of modules designed for AST-driven code generation.

## Global Variables

- `_CURRENT_FILE`: Absolute path to the currently executing Lua script.
- `_FILE_EXTENSION`: Extension of the currently executing script (including the dot).

## Global Functions

### `async(fn, ...)`

Starts a new asynchronous task.

**Parameters:**

- `fn` (function): The function to run asynchronously.
- `...` (any): Arguments to pass to the function.


**Returns:**

- userdata A job handle.

### `await(job)`

Pauses until a job completes.

**Parameters:**

- `job` (userdata): The job handle to wait for.


**Returns:**

- ... any The results of the async function.

### `sleep(seconds)`

Pauses execution for a duration.

**Parameters:**

- `seconds` (number): The number of seconds to sleep.

### `write(id, content)`

Buffers text for a specific output block.

**Parameters:**

- `id` (string): The block ID.
- `content` (string): The content to write.

## Modules

### cogeni

#### `cogeni.get_grammar(ext)`

Looks up a grammar name by file extension.

**Parameters:**

- `ext` (string): The file extension (including dot).


**Returns:**

- string The grammar name.

#### `cogeni.outfile(id, path)`

Redirects write(id) to a file.

**Parameters:**

- `id` (string): The block ID.
- `path` (string): The file path to overwrite.

#### `cogeni.outtag(id, path, tag)`

Redirects write(id) to a specific tag in a file.

**Parameters:**

- `id` (string): The block ID.
- `path` (string): The file path.
- `tag` (string): The tag in the file.

#### `cogeni.process(path)`

Triggers the generation lifecycle for a file.

**Parameters:**

- `path` (string): The file path.


**Returns:**

- boolean True on success.

#### `cogeni.read_ast(source, language)`

Parses source code into a detailed AST.

**Parameters:**

- `source` (string|table): File path or handle.
- `language` (string|nil): The grammar name (optional, defaults to file extension).


**Returns:**

- table The AST table.

#### `cogeni.register_grammar(name, url, opts)`

Registers a custom grammar source URL.

**Parameters:**

- `name` (string): The grammar name.
- `url` (string): The git repository URL.
- `opts` (table): Options (branch, build_cmd, artifact).

### fs

#### `fs.basedir(path)`

Returns the directory of a path.

**Parameters:**

- `path` (string): The file path.


**Returns:**

- string The directory part of the path.

#### `fs.basename(path)`

Returns the last element of a path.

**Parameters:**

- `path` (string): The file path.


**Returns:**

- string The file name.

#### `fs.find(dir, options)`

Recursively finds files and directories.

**Parameters:**

- `dir` (string): The starting directory.
- `options` (table): Filters (type="f"|"d", name="glob", maxdepth=N).


**Returns:**

- table A table where keys are paths and values are 'true'.

#### `fs.join(...)`

Joins multiple path elements.

**Parameters:**

- `...` (string): Path elements to join.


**Returns:**

- string The joined path.

### jq

#### `jq.query(val, query)`

Executes a JQ query against a Lua value.

**Parameters:**

- `val` (any): The value to query (usually an AST table).
- `query` (string): The JQ query string.


**Returns:**

- any|table A single result or a table of results, or nil.

### json

#### `json.decode(str)`

Decodes a JSON string into a Lua table.

**Parameters:**

- `str` (string): The JSON string.


**Returns:**

- any The decoded Lua value.

#### `json.encode(val, options)`

Encodes a Lua value to a JSON string.

**Parameters:**

- `val` (any): The value to encode.
- `options` (table): Optional formatting options (e.g. `{indent=true}`).


**Returns:**

- string The JSON string.

### toml

#### `toml.decode(str)`

Decodes a TOML string into a Lua table.

**Parameters:**

- `str` (string): The TOML string to decode.


**Returns:**

- any The decoded Lua value.

#### `toml.encode(data)`

Encodes a Lua value to a TOML string.

**Parameters:**

- `data` (any): The Lua value to encode.


**Returns:**

- string The TOML string.

### xml

#### `xml.decode(str)`

Decodes an XML string into a Lua table.

**Parameters:**

- `str` (string): The XML string to decode.


**Returns:**

- any The decoded Lua value.

#### `xml.encode(data, options)`

Encodes a Lua value to an XML string.

**Parameters:**

- `data` (any): The Lua value to encode.
- `options` (table|nil): Optional formatting options (e.g. `{root="root_name"}`).


**Returns:**

- string The XML string.

### yaml

#### `yaml.decode(str)`

Decodes a YAML string into a Lua table.

**Parameters:**

- `str` (string): The YAML string to decode.


**Returns:**

- any The decoded Lua value (table, string, number, etc).

#### `yaml.encode(data)`

Encodes a Lua value to a YAML string.

**Parameters:**

- `data` (any): The Lua value to encode.


**Returns:**

- string The YAML string.
