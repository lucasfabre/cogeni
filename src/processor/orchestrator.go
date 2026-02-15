// Package processor manages the orchestration and execution of code generation tasks.
// It implements the core generation lifecycle: extracting blocks from source,
// executing them in a Lua sandbox, and replacing results back into target files.
// It also handles parallel execution, dependency tracking, and atomic filesystem commits.
package processor

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasfabre/codegen/src/config"
)

// Coordinator manages the parallel execution of code generation tasks.
// It ensures that files are processed only once per execution run and
// handles inter-file dependencies automatically when one file's generation
// script reads another file that is also being generated.
type Coordinator struct {
	// cfg is the global application configuration.
	cfg *config.Config
	// tasks tracks all files being processed. Key is absolute path.
	tasks sync.Map // map[string]*Task
	// wg is used to wait for all concurrent worker goroutines to finish.
	wg sync.WaitGroup
	// CleanMode, if true, indicates that blocks should be emptied instead of generated.
	CleanMode bool

	// Results maps absolute file paths to their pending content changes.
	// This buffer allows cogeni to handle recursive updates and preview changes (diff)
	// without actually touching the disk until Commit() is called.
	Results sync.Map // map[string]string

	// runtimePool manages reusable LuaRuntime instances to avoid overhead.
	runtimePool *RuntimePool

	// fileCache caches original file contents to avoid repeated disk I/O.
	// Key is absolute file path, value is the file content as bytes.
	fileCache sync.Map // map[string][]byte
}

// Task represents the processing state of a single file.
type Task struct {
	// path is the absolute path of the file being processed.
	path string
	// done is closed when the file processing is finished.
	done chan struct{}
	// dependencies lists absolute paths of files this task has read during its execution.
	dependencies []string
	// mu protects the dependencies slice.
	mu sync.Mutex
}

// addDependency records that this task depends on the content of another file.
func (t *Task) addDependency(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dependencies = append(t.dependencies, path)
}

// NewCoordinator creates a new orchestrator instance.
func NewCoordinator(cfg *config.Config) *Coordinator {
	return &Coordinator{
		cfg:         cfg,
		runtimePool: NewRuntimePool(cfg, 10), // Pool up to 10 runtimes
	}
}

// RegisterTask manually registers a file to be tracked by the coordinator.
// This is useful for entry point scripts that might not be processed via Process().
func (c *Coordinator) RegisterTask(path string) {
	absPath, _ := filepath.Abs(path)
	c.tasks.LoadOrStore(absPath, &Task{
		path:         absPath,
		done:         make(chan struct{}),
		dependencies: make([]string, 0),
	})
}

// Process initiates the code generation process for a file.
// If the file is already being processed by another goroutine, it waits for it to complete.
// requestor is the absolute path of the file whose script triggered this process call.
func (c *Coordinator) Process(path string, requestor string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for %s: %w", path, err)
	}

	// If this call comes from another task, record the dependency.
	if r, ok := c.tasks.Load(requestor); ok {
		r.(*Task).addDependency(absPath)
	}

	task := &Task{
		path:         absPath,
		done:         make(chan struct{}),
		dependencies: make([]string, 0),
	}

	// Atomically check if the task is already running or register it.
	actual, loaded := c.tasks.LoadOrStore(absPath, task)
	if loaded {
		// Wait for the existing worker to finish.
		<-actual.(*Task).done
		return nil
	}

	// We are the designated worker for this file.
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(task.done)
		if err := ProcessFile(c.cfg, absPath, c); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", absPath, err)
		}
	}()

	return nil
}

// Wait blocks until all registered and spawned tasks have completed.
func (c *Coordinator) Wait() {
	c.wg.Wait()
	// Close the runtime pool to free resources
	c.runtimePool.Close()
}

// Commit flushes all buffered changes from the Results map to the filesystem.
// It only writes files that have actually changed to minimize disk I/O.
// Uses content hashing for efficient change detection on large files.
func (c *Coordinator) Commit() error {
	var errs []error
	c.Results.Range(func(key, value any) bool {
		path := key.(string)
		content := value.(string)

		// Optimization: skip write if content is identical to what's on disk.
		// Use SHA256 hashing for efficient comparison of large files.
		original, err := os.ReadFile(path)
		if err == nil {
			originalHash := sha256.Sum256(original)
			newHash := sha256.Sum256([]byte(content))
			if originalHash == newHash {
				return true // No change, skip write
			}
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			errs = append(errs, fmt.Errorf("failed to write %s: %w", path, err))
		}
		return true
	})

	if len(errs) > 0 {
		return fmt.Errorf("commit failed with %d errors: %v", len(errs), errs)
	}
	return nil
}

// GetResults returns a snapshot of all pending file changes.
func (c *Coordinator) GetResults() map[string]string {
	res := make(map[string]string)
	c.Results.Range(func(key, value any) bool {
		res[key.(string)] = value.(string)
		return true
	})
	return res
}

// GetResult retrieves the current buffered content for a file path.
// It returns (content, true) if the file has been processed, or ("", false) otherwise.
func (c *Coordinator) GetResult(path string) (string, bool) {
	absPath, _ := filepath.Abs(path)
	val, ok := c.Results.Load(absPath)
	if !ok {
		return "", false
	}
	return val.(string), true
}

// GetGraph returns the recorded file dependency graph.
// Map key is the file path, value is the list of files it depends on.
func (c *Coordinator) GetGraph() map[string][]string {
	graph := make(map[string][]string)
	c.tasks.Range(func(key, value any) bool {
		task := value.(*Task)
		if len(task.dependencies) > 0 {
			graph[task.path] = task.dependencies
		}
		return true
	})
	return graph
}

// WaitForReader ensures that a task waits for a file to be fully updated before reading it.
// This is critical for inter-file dependencies to avoid reading stale or partially written content.
func (c *Coordinator) WaitForReader(path string, requestor string) error {
	absPath, _ := filepath.Abs(path)

	if absPath == requestor {
		// Self-dependency is ignored to prevent immediate deadlocks.
		return nil
	}

	// Record the dependency.
	if r, ok := c.tasks.Load(requestor); ok {
		r.(*Task).addDependency(absPath)
	}

	if t, ok := c.tasks.Load(absPath); ok {
		task := t.(*Task)
		// Wait for the file's worker to signal completion.
		<-task.done
	}
	return nil
}

// readFileContent reads file content from cache or disk.
// This reduces disk I/O for files that are read multiple times.
func (c *Coordinator) readFileContent(path string) ([]byte, error) {
	absPath, _ := filepath.Abs(path)

	// Check cache first
	if cached, ok := c.fileCache.Load(absPath); ok {
		return cached.([]byte), nil
	}

	// Read from disk
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Cache the content
	c.fileCache.Store(absPath, content)
	return content, nil
}
