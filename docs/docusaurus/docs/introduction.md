---
sidebar_position: 1
---

# Introduction

`cogeni` is a language-agnostic runtime for generating reproducible artifacts from source code and machine-readable specs. It uses [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) for AST parsing, [Lua](https://www.lua.org/) for programmable generation logic, and a dependency-aware execution engine to keep derived outputs in sync.

## Project Overview

The core idea of `cogeni` is to turn source-of-truth inputs into repeatable derived outputs. Those inputs can be code, structured files, or external specs. `cogeni` parses code into a queryable AST, exposes that structure inside Lua, and lets you generate artifacts such as:

- documentation and reference material
- SDKs, CLIs, and server/client code
- schemas, config files, and synchronized generated sections

Agents are useful for writing and maintaining Lua or JQ rules. `cogeni` provides the deterministic runtime that executes those rules repeatedly in local workflows, watch mode, and CI.

### Key Features
- **Language Agnostic**: Works with any language that has a Tree-sitter grammar.
- **Artifact-Oriented**: Generate docs, SDKs, CLIs, schemas, config, and code from a shared runtime.
- **AST-Aware**: Access source structure precisely instead of relying on regex-based parsing.
- **Programmable With Lua**: Define transformations in a small, flexible scripting layer.
- **JQ Integration**: Query ASTs, JSON, and decoded structured inputs with familiar JQ expressions.
- **Dependency-Aware Execution**: Track file relationships and process generation tasks in parallel.
- **Plugin-Friendly Core**: Use the core as a base for reusable Lua libraries and domain-specific plugins.

## Architecture

The project is structured into several Go packages:

- `src/cmd/`: CLI implementation using [Cobra](https://github.com/spf13/cobra).
- `src/astparser/`: Manages Tree-sitter grammars and transforms native ASTs into a serializable format for Lua.
- `src/lua_runtime/`: Provides a sandboxed Lua environment with custom bindings for AST access, file system, JQ, and structured data helpers.
- `src/processor/`: Orchestrates the generation lifecycle, handling file dependencies and concurrent tasks.
- `src/config/`: Configuration management.

The long-term shape of `cogeni` is a generic core plus reusable Lua libraries and plugins. The core handles parsing, orchestration, and output management. Plugins can focus on opinionated workflows such as OpenAPI-driven SDKs, generated docs, or CLI scaffolding.

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

## Typical Workflows

- Generate docs from source code, comments, or specs.
- Generate SDKs, CLIs, and stubs from OpenAPI or similar inputs.
- Keep generated sections inside existing files synchronized with a canonical source.
- Build project-specific automation as reusable Lua plugins.
