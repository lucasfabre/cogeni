---
sidebar_position: 5
title: "CLI Usage"
---

# CLI Usage

`cogeni` provides a set of subcommands to manage the code generation process.

## Subcommands

### `run [file...]`
Executes Lua scripts or processes files with embedded `cogeni` blocks.

```bash
cogeni run script.lua
cogeni run models.py
```

### `clean [file...]`
Removes generated content from files, reverting them to their template state.

```bash
cogeni clean generated/
```

### `diff [file...]`
Shows a diff between the current file content and what `cogeni` would generate.

```bash
cogeni diff models.py
```

### `ast <file>`
Prints the JSON-serialized AST of the given file. Useful for debugging AST queries.

```bash
cogeni ast src/main.py
```

### `show dependency-graph <file>`
Visualizes the file-level dependencies calculated by the orchestrator.

```bash
cogeni show dependency-graph src/main.py
```

### `repl`
Starts an interactive Lua REPL with all `cogeni` modules pre-loaded.

```bash
cogeni repl
```

### `config`
Displays or manages the current configuration.

```bash
cogeni config
```

## Configuration

`cogeni` looks for a `config.yaml` file in standard XDG directories (e.g., `~/.config/cogeni/config.yaml`).

```yaml
grammar:
  location: "~/.local/share/cogeni/grammars"
  mapping:
    ".py": "python"
    ".ts": "typescript"
  sources:
    python:
      url: "https://github.com/tree-sitter/tree-sitter-python"
```
