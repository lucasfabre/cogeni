# Grammar Configuration

`cogeni` relies on [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) grammars to parse source code. By default, it supports Python and TypeScript, but you can configure additional languages or override the defaults.

## Configuration File

The configuration is loaded from `config.yaml` in standard XDG directories (e.g., `~/.config/cogeni/config.yaml`).

## Adding a New Language

To add support for a new language, you need to provide the grammar source and map file extensions to the language name.

```yaml
grammar:
  location: "~/.local/share/cogeni/grammars"
  mapping:
    ".go": "go"
    ".rs": "rust"
  sources:
    go:
      url: "https://github.com/tree-sitter/tree-sitter-go"
      branch: "master"
    rust:
      url: "https://github.com/tree-sitter/tree-sitter-rust"
```

## Overriding Defaults

If you want to use a specific version or fork of a grammar, you can override the default source URL.

```yaml
grammar:
  sources:
    python:
      url: "https://github.com/my-org/tree-sitter-python"
```

## Managing Grammars

`cogeni` will automatically download and compile grammars the first time they are needed. You can manually trigger an update or check the status using the CLI (future feature).

## Troubleshooting

If `cogeni` fails to parse a file, ensure that:
1. The file extension is correctly mapped in `config.yaml`.
2. The grammar source URL is reachable.
3. You have a C compiler installed (required for building grammars).
