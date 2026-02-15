package cmd

import (
	"path/filepath"

	luaruntime "github.com/lucasfabre/codegen/src/lua_runtime"
	"github.com/lucasfabre/codegen/src/processor"
)

// executionContext encapsulates the shared state and logic required for most cogeni subcommands.
// It manages the relationship between the central Coordinator (task management)
// and the LuaRuntime (script execution environment).
type executionContext struct {
	// coordinator manages the execution graph and filesystem buffers.
	coordinator *processor.Coordinator
	// runtime provides the environment for executing Lua-based orchestration or generation.
	runtime *luaruntime.LuaRuntime
}

// newExecutionContext creates a unified context for executing cogeni operations.
// It automatically configures the Lua runtime with the necessary hooks to
// interact with the coordinator for recursive processing and inter-file dependency waiting.
func newExecutionContext() (*executionContext, error) {
	coordinator := processor.NewCoordinator(cfg)

	rt, err := luaruntime.New(cfg)
	if err != nil {
		return nil, err
	}

	// Link the runtime hooks back to the coordinator.
	rt.ProcessFunc = func(path, requestor string) error {
		return coordinator.Process(path, requestor)
	}
	rt.WaitFunc = func(path, requestor string) error {
		return coordinator.WaitForReader(path, requestor)
	}
	rt.ReadFunc = coordinator.GetResult

	return &executionContext{
		coordinator: coordinator,
		runtime:     rt,
	}, nil
}

// runEntrypoint executes an orchestration script (typically cogeni.lua) as the starting point.
// It handles task registration, script execution, result capturing, and waiting for completion.
func (ctx *executionContext) runEntrypoint(scriptPath string) error {
	absPath, _ := filepath.Abs(scriptPath)

	// Manually register the entry point script as a task so it can track its own dependencies.
	ctx.coordinator.RegisterTask(absPath)

	if err := ctx.runtime.ExecuteFile(absPath); err != nil {
		return err
	}

	// Capture any write() calls made directly in the entry point script.
	if err := ctx.coordinator.CaptureResults(ctx.runtime, absPath); err != nil {
		return err
	}

	// Block until all recursively triggered Process() calls are finished.
	ctx.coordinator.Wait()

	// Finalize any async tasks that might be pending in the Lua engine.
	ctx.runtime.Schedule()

	return nil
}
