# High-Performance Stress Test

This directory contains a self-generating stress test environment for `cogeni`. It is designed to push the parallel execution engine, AST parser, and Lua runtime to their limits for performance profiling.

## Usage

### 1. Generate the Environment

Run the generator script to create 1,000+ Python files with inter-dependencies and accompanying JSON configuration files:

```bash
./cogeni run examples/performance/generate.lua
```

This will create a `src/` directory inside `examples/performance/`.

### 2. Run the Stress Test

Run the orchestration script to process all generated files in parallel:

```bash
./cogeni run examples/performance/cogeni.lua
```

This script:
- Finds all generated Python files.
- Spawns asynchronous tasks to process each file using `cogeni.process`.
- Each file parses its own AST (using Tree-sitter) and reads its config file.
- Runs a concurrent aggregation task to sum IDs from all JSON files.

### 3. Profiling

To generate a flame graph for performance analysis, use the provided `Justfile` target:

```bash
just profile-perf
```

This will:
1.  Build `cogeni`.
2.  Generate the stress test environment.
3.  Run the stress test with CPU profiling enabled.
4.  Open the `pprof` web interface to view the flame graph.

**Note:** You need `go tool pprof` installed.

## What it tests

- **Concurrency**: Spawns 1,000+ goroutines/Lua tasks.
- **Tree-sitter**: Parses 1,000+ Python files concurrently.
- **Lua Interop**: Heavy use of `cogeni` module, `fs`, `json`, and `jq`.
- **File I/O**: Reads and writes thousands of files.
- **Dependency Graph**: Random imports create a complex dependency web.
