---
sidebar_position: 4
title: "Parallel Processing"
---

# Parallel Processing

`cogeni` is designed to be highly concurrent, handling file dependencies efficiently.

## How it Works

1. **Dependency Analysis**: `cogeni` analyzes the dependency graph of your source files. It detects imports and references to determine the order of processing.
2. **Parallel Execution**: Files that are independent of each other are processed in parallel.
3. **Synchronization**: Dependent files wait for their dependencies to finish processing.

## Benefits

- **Speed**: Utilize multiple CPU cores to generate code faster.
- **Correctness**: Ensure that generated code is always up-to-date with its dependencies.

## Controlling Concurrency

By default, `cogeni` runs up to 10 concurrent processing jobs. You can control the maximum number of parallel tasks and Lua runtimes using the `-j` or `--jobs` flag, the `COGENI_CONCURRENCY` environment variable, or by setting `concurrency` in your `config.yaml`. The underlying Go runtime concurrency can also be bounded with the standard `GOMAXPROCS` environment variable.

```bash
# Limit to 4 concurrent jobs
cogeni run -j 4
# Or use the environment variable
COGENI_CONCURRENCY=4 cogeni run
```

## Circular Dependencies

`cogeni` detects circular dependencies and will halt execution with an error message, preventing infinite loops. Use `cogeni show dependency-graph` to visualize the cycle.

## Best Practices

- Avoid deeply nested dependencies if possible.
- Use explicit imports to clarify dependencies.
- If you encounter build order issues, inspect the dependency graph.
