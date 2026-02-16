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

By default, `cogeni` uses all available CPU cores. You can limit the concurrency using the `-j` flag (future feature) or by setting an environment variable `GOMAXPROCS`.

```bash
# Limit to 4 cores
GOMAXPROCS=4 cogeni run
```

## Circular Dependencies

`cogeni` detects circular dependencies and will halt execution with an error message, preventing infinite loops. Use `cogeni show dependency-graph` to visualize the cycle.

## Best Practices

- Avoid deeply nested dependencies if possible.
- Use explicit imports to clarify dependencies.
- If you encounter build order issues, inspect the dependency graph.
