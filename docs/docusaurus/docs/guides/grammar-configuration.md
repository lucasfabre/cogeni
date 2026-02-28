---
sidebar_position: 2
---

# Grammar Configuration

`cogeni` uses Tree-sitter grammars to parse source code into an AST. By default, it manages downloading and building grammars for you. This guide explains how grammar resolution works and how you can configure custom grammars.

## How `cogeni` Fetches Grammars

When `cogeni` encounters a file it needs to parse (e.g., via `cogeni.read_ast(path, "python")`), it first checks if the grammar for that language is already available locally in its cache directory (default is `~/.cache/cogeni/grammars`).

If the grammar is not found, `cogeni` will attempt to automatically download and compile it from standard Tree-sitter GitHub repositories (e.g., `https://github.com/tree-sitter/tree-sitter-python`).

## Registering Custom Grammars

You may need to parse languages that `cogeni` does not know about by default, or you might want to use a specific fork or branch of a grammar. You can register custom grammars in two ways:

### 1. Via `config.yaml`

You can define custom grammars in your `config.yaml` file (usually located at `~/.config/cogeni/config.yaml` or `.cogeni.yaml` in your project root).

```yaml
grammars:
  - name: mylang
    url: https://github.com/myuser/tree-sitter-mylang
    branch: main
    build_cmd: "make"
    artifact: "mylang.so"
    extensions:
      - ".mlg"
      - ".my"
```

### 2. Via Lua Script

You can also dynamically register grammars within your Lua scripts before they are needed:

```lua
cogeni.register_grammar("mylang", "https://github.com/myuser/tree-sitter-mylang", {
    branch = "main",
    build_cmd = "make",
    artifact = "mylang.so"
})

-- Now you can parse files with your custom grammar
local ast = cogeni.read_ast("example.mlg", "mylang")
```

## Manual Grammar Compilation

If you are working in an environment without internet access or simply prefer to manage grammars manually, you can compile the grammar yourself and place it in the cache directory.

1. Clone the grammar repository.
2. Build it according to its instructions (typically `make` or using a C compiler to create a shared object `.so` / `.dylib` / `.dll`).
3. Copy the compiled artifact to the `cogeni` grammar cache directory, naming it `grammar-<name>.so` (e.g., `~/.cache/cogeni/grammars/grammar-mylang.so`).
