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

	"github.com/lucasfabre/cogeni/src/config"
)

// Coordinator manages the parallel execution of code generation tasks.
// It ensures that files are processed only once per execution run and
// handles inter-file dependencies automatically.
type Coordinator struct {
	cfg       *config.Config
	mu        sync.RWMutex
	tasks     map[string]*Task
	wg        sync.WaitGroup
	CleanMode bool

	// Results buffers content changes to handle recursive updates and preview changes.
	Results map[string]string

	runtimePool *RuntimePool

	fileCache map[string][]byte
}

// Task represents the processing state of a single file.
type Task struct {
	path         string
	done         chan struct{}
	dependencies []string
	mu           sync.Mutex
}

// addDependency records that this task depends on the content of another file.
func (t *Task) addDependency(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dependencies = append(t.dependencies, path)
}

// NewCoordinator creates a new Coordinator.
func NewCoordinator(cfg *config.Config) *Coordinator {
	concurrency := 10
	if cfg != nil && cfg.Concurrency > 0 {
		concurrency = cfg.Concurrency
	}

	return &Coordinator{
		cfg:         cfg,
		tasks:       make(map[string]*Task),
		Results:     make(map[string]string),
		fileCache:   make(map[string][]byte),
		runtimePool: NewRuntimePool(cfg, concurrency), // Pool sized by configuration
	}
}

// RegisterTask manually registers a file to be tracked by the coordinator.
func (c *Coordinator) RegisterTask(path string) {
	absPath, _ := filepath.Abs(path)
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.tasks[absPath]; !ok {
		c.tasks[absPath] = &Task{
			path:         absPath,
			done:         make(chan struct{}),
			dependencies: make([]string, 0),
		}
	}
}

// FinishTask marks a manually registered task as complete.
func (c *Coordinator) FinishTask(path string) {
	absPath, _ := filepath.Abs(path)
	c.mu.RLock()
	task, ok := c.tasks[absPath]
	c.mu.RUnlock()

	if ok {
		select {
		case <-task.done:
			// Already closed
		default:
			close(task.done)
		}
	}
}

// Process initiates the code generation process for a file.
// If the file is already being processed by another goroutine, it waits for it to complete.
func (c *Coordinator) Process(path string, requestor string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for %s: %w", path, err)
	}

	// If this call comes from another task, record the dependency.
	c.mu.RLock()
	requestorTask, ok := c.tasks[requestor]
	c.mu.RUnlock()

	if ok {
		requestorTask.addDependency(absPath)
	}

	c.mu.Lock()
	task, loaded := c.tasks[absPath]
	if !loaded {
		task = &Task{
			path:         absPath,
			done:         make(chan struct{}),
			dependencies: make([]string, 0),
		}
		c.tasks[absPath] = task
	}
	c.mu.Unlock()

	if loaded {
		// Wait for the existing worker to finish.
		<-task.done
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
	c.runtimePool.Close()
}

// WaitWithoutClosing blocks until all registered and spawned tasks have completed,
// but keeps the runtime pool open for subsequent runs.
func (c *Coordinator) WaitWithoutClosing() {
	c.wg.Wait()
}

// Close releases resources held by the coordinator.
func (c *Coordinator) Close() {
	c.runtimePool.Close()
}

// Invalidate invalidates a file and all its dependents, removing them from the processing cache.
// This forces them to be re-processed in the next run.
func (c *Coordinator) Invalidate(changedPath string) {
	absPath, _ := filepath.Abs(changedPath)
	visited := make(map[string]bool)
	queue := []string{absPath}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if visited[curr] {
			continue
		}
		visited[curr] = true

		// Clear state for the invalidated file
		c.mu.Lock()
		delete(c.fileCache, curr)
		delete(c.Results, curr)
		delete(c.tasks, curr)

		dependents := make([]string, 0)
		for _, task := range c.tasks {
			task.mu.Lock()
			depends := false
			for _, dep := range task.dependencies {
				if dep == curr {
					depends = true
					break
				}
			}
			task.mu.Unlock()

			if depends {
				dependents = append(dependents, task.path)
			}
		}
		c.mu.Unlock()

		queue = append(queue, dependents...)
	}
}

// Commit flushes all buffered changes from the Results map to the filesystem.
// It only writes files that have actually changed to minimize disk I/O.
// Uses content hashing for efficient change detection on large files.
func (c *Coordinator) Commit() error {
	var errs []error

	c.mu.RLock()
	// Copy map to avoid holding lock during I/O
	resultsCopy := make(map[string]string)
	for k, v := range c.Results {
		resultsCopy[k] = v
	}
	c.mu.RUnlock()

	for path, content := range resultsCopy {
		// Optimization: skip write if content is identical to what's on disk.
		// Use SHA256 hashing for efficient comparison of large files.
		original, err := os.ReadFile(path)
		if err == nil {
			originalHash := sha256.Sum256(original)
			newHash := sha256.Sum256([]byte(content))
			if originalHash == newHash {
				continue // No change, skip write
			}
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			errs = append(errs, fmt.Errorf("failed to write %s: %w", path, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("commit failed with %d errors: %v", len(errs), errs)
	}
	return nil
}

// GetResults returns a snapshot of all pending file changes.
func (c *Coordinator) GetResults() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make(map[string]string)
	for key, value := range c.Results {
		res[key] = value
	}
	return res
}

// GetResult retrieves the current buffered content for a file path.
// It returns (content, true) if the file has been processed, or ("", false) otherwise.
func (c *Coordinator) GetResult(path string) (string, bool) {
	absPath, _ := filepath.Abs(path)
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Results[absPath]
	return val, ok
}

// GetGraph returns the recorded file dependency graph.
// Map key is the file path, value is the list of files it depends on.
func (c *Coordinator) GetGraph() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	graph := make(map[string][]string)
	for key, task := range c.tasks {
		graph[key] = task.dependencies
	}
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
	c.mu.RLock()
	requestorTask, ok := c.tasks[requestor]
	c.mu.RUnlock()

	if ok {
		requestorTask.addDependency(absPath)
	}

	c.mu.RLock()
	task, ok := c.tasks[absPath]
	c.mu.RUnlock()

	if ok {
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
	c.mu.RLock()
	cached, ok := c.fileCache[absPath]
	c.mu.RUnlock()

	if ok {
		return cached, nil
	}

	// Read from disk
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Cache the content
	c.mu.Lock()
	c.fileCache[absPath] = content
	c.mu.Unlock()

	return content, nil
}

// DetectCycles checks for circular dependencies in the task graph.
// Returns an error describing the cycle if one is found.
func (c *Coordinator) DetectCycles() error {
	graph := c.GetGraph()
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var checkCycle func(node string) ([]string, bool)
	checkCycle = func(node string) ([]string, bool) {
		visited[node] = true
		recursionStack[node] = true

		if deps, ok := graph[node]; ok {
			for _, dep := range deps {
				if !visited[dep] {
					if path, found := checkCycle(dep); found {
						return append([]string{node}, path...), true
					}
				} else if recursionStack[dep] {
					return []string{node, dep}, true
				}
			}
		}

		recursionStack[node] = false
		return nil, false
	}

	for node := range graph {
		if !visited[node] {
			if path, found := checkCycle(node); found {
				cwd, _ := os.Getwd()
				var formattedPath string
				for i, p := range path {
					rel, err := filepath.Rel(cwd, p)
					if err != nil {
						rel = p
					}
					if i > 0 {
						formattedPath += " -> "
					}
					formattedPath += rel
				}
				return fmt.Errorf("circular dependency detected: %s", formattedPath)
			}
		}
	}

	return nil
}
