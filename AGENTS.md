# cogeni - Language-Agnostic Code Generation Tool

`cogeni` is a powerful, language-agnostic code generation tool that leverages [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) for AST parsing and [Lua](https://www.lua.org/) for generation logic. It allows developers to programmatically transform and generate source code across multiple languages using a unified scripting interface.

## Project Overview

The core idea of `cogeni` is to parse source files into a rich, queryable Abstract Syntax Tree (AST), expose this AST to a Lua environment, and then use Lua scripts to generate new code or modify existing files. It handles complex tasks like dependency tracking between generated files and parallel execution.

### Key Features
- **Language Agnostic**: Supports any language with a Tree-sitter grammar.
- **Lua Scripting**: Use the full power of Lua to define your code generation logic.
- **AST-Aware**: Access detailed source code structure, not just regex-based matching.
- **JQ Integration**: Built-in support for querying JSON data using JQ within Lua.
- **Parallel Orchestration**: Automatically manages file dependencies and processes files in parallel.
- **Embedded FS and JSON**: Custom Lua modules for file system operations and JSON manipulation.

## Architecture

The project is structured into several Go packages:

- `src/cmd/`: CLI implementation using [Cobra](https://github.com/spf13/cobra).
- `src/astparser/`: Manages Tree-sitter grammars and transforms native ASTs into a serializable format for Lua.
- `src/lua_runtime/`: Provides a sandboxed Lua environment with custom bindings for AST access, file system, JQ, and JSON.
- `src/processor/`: Orchestrates the code generation lifecycle, handling file dependencies and concurrent tasks.
- `src/config/`: Configuration management using [Viper](https://github.com/spf13/viper).

## Technologies

- **Go 1.25+**: Primary implementation language.
- **Lua (gopher-lua)**: Embedded scripting engine.
- **Tree-sitter (go-tree-sitter)**: Incremental parsing system.
- **JQ (gojq)**: Pure Go implementation of JQ.
- **Just**: Task runner for development commands.

## Getting Started

### Prerequisites
- Go 1.25 or higher.
- [Just](https://github.com/casey/just) task runner (optional, but recommended).
- [Mise](https://mise.jdx.dev/) for tool management (optional).

### Building and Running
```bash
# Build the cogeni binary
just build

# Run cogeni (it will look for cogeni.lua in the current directory)
./cogeni

# Run a specific Lua script
./cogeni run script.lua
```

## Subcommands

- `run [file...]`: Executes Lua scripts or processes files with embedded `cogeni` blocks.
- `clean [file...]`: Removes generated content from files, reverting them to their template state.
- `diff [file...]`: Shows a diff between the current file content and what `cogeni` would generate.
- `ast <file>`: Prints the JSON-serialized AST of the given file.
- `show dependency-graph <file>`: Visualizes the file-level dependencies calculated by the orchestrator.
- `repl`: Starts an interactive Lua REPL with all `cogeni` modules pre-loaded.
- `config`: Displays or manages the current configuration.

## Standalone Mode

In addition to dedicated `.lua` scripts, `cogeni` can process files with embedded Lua blocks. These blocks are defined within comments:

```python
# <cogeni>
#   cogeni.outfile("models", "generated_models.py")
#   write("models", "class User: pass")
# </cogeni>
```

## Examples

- **[Full-stack SaaS](./examples/fullstack)**: A comprehensive example demonstrating recursive file processing, AST parsing of Python models, and generation of a FastAPI backend and TypeScript SDK.

### Development Commands
```bash
# Run all tests (Go unit tests and shell integration tests)
just test

# Run Lua-specific integration tests
just test-lua

# Lint the codebase
just lint

# Format the codebase
just fmt
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

## Lua API

`cogeni` exposes several global modules to the Lua environment:

- `cogeni`: Core module for reading ASTs and registering grammars.
  - `read_ast(path)`: Parses a file and returns its AST.
  - `process(path)`: Recursively triggers processing for another file.
  - `outfile(id, path)`: Directs `write(id, content)` output to a whole file.
  - `outtag(id, path, tag)`: Directs `write(id, content)` output to a tagged block in a file.
- `fs`: File system utilities (`find`, `join`, `basedir`, `basename`).
- `json`: JSON encoding and decoding.
- `jq`: Execute JQ queries on Lua tables/JSON strings.
- `write(id, content)`: Global function to buffer generated content for a specific target.

## Testing Practices

- **Go Tests**: Located in `src/` alongside implementation (e.g., `config_test.go`) and in `tests/`.
- **Lua Tests**: Located in `tests/lua/`, used to verify Lua bindings and generation logic.
- **Shell Tests**: Located in `tests/shell/`, used for end-to-end CLI integration testing.
