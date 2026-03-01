---
sidebar_position: 1
---

# Introduction

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
- `src/config/`: Configuration management.

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
