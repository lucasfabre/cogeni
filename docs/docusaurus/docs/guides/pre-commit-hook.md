---
sidebar_position: 5
title: "Pre-commit Hook"
---

# Using the Pre-commit Hook

`cogeni` provides a pre-commit hook to ensure that your generated code is always up-to-date with your source files. This prevents outdated generated files from being committed to your repository.

## Installation

Add the following to your `.pre-commit-config.yaml` file:

```yaml
- repo: https://github.com/lucasfabre/codegen
  rev: v0.1.0  # Use the latest release tag
  hooks:
    - id: cogeni
```

Make sure you have `cogeni` installed locally and available in your `PATH`. The hook script checks for `cogeni` presence before running.

## Configuration

By default, the hook runs `cogeni run` without arguments, which executes all generators in the current directory (or as defined by `cogeni.lua` / `cogeni.toml`).

If you need to pass specific arguments to `cogeni run`, you can use the `args` key:

```yaml
- repo: https://github.com/lucasfabre/codegen
  rev: v0.1.0
  hooks:
    - id: cogeni
      args: ["config/cogeni.lua", "--clean"]
```

## Behavior

When you run `git commit`, the hook will:

1.  Check if `cogeni` is installed.
2.  Run `cogeni run` (with any configured arguments).
3.  Check if any files were modified by the run.

If files were modified (meaning the generated code was out-of-date), the commit will fail with a message:

```
Files were modified by cogeni. Please stage the changes and commit again.
```

This ensures that you review the changes made by the generator before committing them. Simply stage the modified files (`git add ...`) and run `git commit` again.
