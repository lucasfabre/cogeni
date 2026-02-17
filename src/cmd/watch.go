package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	luaruntime "github.com/lucasfabre/codegen/src/lua_runtime"
	"github.com/lucasfabre/codegen/src/processor"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [script.lua]",
	Short: "Watch files and re-run generation on changes",
	Long:  `Watch the entry script and all its dependencies, re-running the generation process whenever a file changes.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer watcher.Close()

		coordinator := processor.NewCoordinator(cfg)
		defer coordinator.Close()

		// Determine the entry point (file or stdin)
		var entryScript string
		var entryPath string
		var isStdin bool

		if len(args) > 0 {
			entryPath, _ = filepath.Abs(args[0])
		} else {
			// Stdin
			fi, _ := os.Stdin.Stat()
			if (fi.Mode() & os.ModeCharDevice) == 0 {
				scanner := bufio.NewScanner(os.Stdin)
				var script strings.Builder
				for scanner.Scan() {
					script.WriteString(scanner.Text())
					script.WriteString("\n")
				}
				entryScript = script.String()
				isStdin = true
				entryPath = "stdin" // Dummy path for dependency tracking
			} else {
				return cmd.Help()
			}
		}

		// Keep track of what we are watching to avoid duplicate Adds
		watchedFiles := make(map[string]bool)
		var watchMutex sync.Mutex

		updateWatchList := func() {
			watchMutex.Lock()
			defer watchMutex.Unlock()

			// Helper to add file to watcher
			addWatch := func(path string) {
				if path == "" || path == "stdin" {
					return
				}
				if !watchedFiles[path] {
					if err := watcher.Add(path); err == nil {
						watchedFiles[path] = true
						fmt.Printf("Watching: %s\n", path)
					}
				}
			}

			// Always watch the entry file if it's a real file
			if !isStdin {
				addWatch(entryPath)
			}

			// Watch all files tracked by the coordinator
			graph := coordinator.GetGraph()
			for taskPath, deps := range graph {
				addWatch(taskPath)
				for _, dep := range deps {
					addWatch(dep)
				}
			}
		}

		// Execution function
		execute := func() error {
			start := time.Now()
			fmt.Printf("[\033[36m%s\033[0m] Starting build...\n", start.Format(time.TimeOnly))

			// Create a new runtime for each execution to ensure fresh state
			rt, err := luaruntime.New(cfg)
			if err != nil {
				return err
			}
			defer rt.Close()

			// Register entry path as a task so dependencies are recorded
			coordinator.RegisterTask(entryPath)
			// Ensure we signal completion of the entry task so dependents don't hang
			defer coordinator.FinishTask(entryPath)

			// Enable recursive processing
			rt.ProcessFunc = func(path, requestor string) error {
				return coordinator.Process(path, requestor)
			}
			rt.WaitFunc = coordinator.WaitForReader
			rt.ReadFunc = coordinator.GetResult

			if !isStdin {
				if err := rt.ExecuteFile(entryPath); err != nil {
					fmt.Printf("Error executing file: %v\n", err)
					return nil // Don't crash on script errors
				}
			} else {
				// Set dummy current file for dependency tracking
				rt.L.PushString(entryPath)
				rt.L.SetGlobal("_CURRENT_FILE")
				if err := rt.DoString(entryScript); err != nil {
					fmt.Printf("Error executing script: %v\n", err)
					return nil
				}
			}

			// Mark entry task as done before waiting for workers (though we deferred it,
			// calling it explicitly here ensures B wakes up if it's waiting)
			coordinator.FinishTask(entryPath)

			coordinator.WaitWithoutClosing()
			rt.Schedule()

			// Capture results from the entry script
			if err := coordinator.CaptureResults(rt, entryPath); err != nil {
				fmt.Printf("Error capturing results: %v\n", err)
			}

			// Check for circular dependencies
			if err := coordinator.DetectCycles(); err != nil {
				fmt.Printf("[\033[31mError\033[0m] %v\n", err)
				return nil // Stop here, do not commit
			}

			if err := coordinator.Commit(); err != nil {
				fmt.Printf("[\033[31mError\033[0m] Failed to commit changes: %v\n", err)
			} else {
				fmt.Printf("[\033[32mSuccess\033[0m] Build completed in %v\n", time.Since(start))
			}

			updateWatchList()

			return nil
		}

		// Initial run
		if err := execute(); err != nil {
			return err
		}

		fmt.Println("Watching for changes. Press 'q' to exit.")

		// Setup signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Setup stop channel for user input
		stopChan := make(chan struct{})

		// Start listening for user input if we're not reading script from stdin
		if !isStdin {
			go func() {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					if strings.TrimSpace(scanner.Text()) == "q" {
						close(stopChan)
						return
					}
				}
			}()
		}

		// Event loop
		for {
			select {
			case <-sigChan:
				fmt.Println("\nReceived interrupt, shutting down...")
				return nil
			case <-stopChan:
				fmt.Println("Exiting watch mode...")
				return nil
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
					// Handle Rename/Remove by clearing watch state so it gets re-added
					if event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
						watchMutex.Lock()
						delete(watchedFiles, event.Name)
						watchMutex.Unlock()
					}

					// Invalidate the cache for the changed file and its dependents
					coordinator.Invalidate(event.Name)

					// Re-run, but delay slightly to allow file system to settle (especially for atomic saves/re-creates)
					time.Sleep(100 * time.Millisecond)
					if err := execute(); err != nil {
						fmt.Printf("Error during re-build: %v\n", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				fmt.Printf("Watcher error: %v\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
