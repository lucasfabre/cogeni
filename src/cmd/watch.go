package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	luaruntime "github.com/lucasfabre/cogeni/src/lua_runtime"
	"github.com/lucasfabre/cogeni/src/processor"
	"github.com/spf13/cobra"
	lua "github.com/yuin/gopher-lua"
)

var watchCmd = &cobra.Command{
	Use:   "watch [script.lua]",
	Short: "Watch files and re-run generation on changes",
	Long:  `Watch the entry script and all its dependencies, re-running the generation process whenever a file changes.`,
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"lua"}, cobra.ShellCompDirectiveFilterFileExt
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		coordinator := processor.NewCoordinator(cfg)
		defer coordinator.Close()

		ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stopSignals()

		var shutdownOnce sync.Once
		shutdown := func(reason string) error {
			var shutdownErr error
			shutdownOnce.Do(func() {
				if reason != "" {
					fmt.Println(reason)
				}
				stopSignals()
				shutdownErr = watcher.Close()
			})
			return shutdownErr
		}
		defer func() { _ = shutdown("") }()

		// Determine the entry point (file or stdin)
		var entryScript string
		var entryPath string
		var isStdin bool

		if len(args) > 0 {
			entryPath, _ = filepath.Abs(args[0])
		} else {
			// Check for default entrypoint first
			if defEntry, ok := findDefaultEntrypoint(); ok {
				entryPath, _ = filepath.Abs(defEntry)
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
						cwd, _ := os.Getwd()
						rel, err := filepath.Rel(cwd, path)
						if err != nil {
							rel = path
						}
						fmt.Printf("Watching: %s\n", rel)
					} else {
						fmt.Printf("[\033[33mWarning\033[0m] Failed to watch %s: %v\n", path, err)
					}
				}
			}

			// Always watch the entry file if it's a real file
			if !isStdin {
				addWatch(entryPath)
			}

			// Watch the source tasks tracked by the coordinator.
			// Generated output files are stored as dependencies too, but watching
			// them causes self-triggered rebuild loops on some platforms.
			graph := coordinator.GetGraph()
			for taskPath := range graph {
				addWatch(taskPath)
			}
		}

		// Execution function
		execute := func() error {
			start := time.Now()
			fmt.Printf("[\033[36m%s\033[0m] Starting build...\n", start.Format(time.TimeOnly))

			// Phase 1: Prepare watch state and runtime
			rt, err := luaruntime.New(cfg)
			if err != nil {
				fmt.Printf("[\033[31mError\033[0m] Failed to initialize runtime: %v\n", err)
				return nil // Don't crash on initialization errors
			}
			defer rt.Close()

			coordinator.RegisterTask(entryPath)
			defer coordinator.FinishTask(entryPath)

			rt.ProcessFunc = func(path, requestor string) error {
				return coordinator.Process(path, requestor)
			}
			rt.WaitFunc = coordinator.WaitForReader
			rt.ReadFunc = coordinator.GetResult

			// Phase 2: Execute entry script
			if !isStdin {
				if err := rt.ExecuteFile(entryPath); err != nil {
					fmt.Printf("[\033[31mError\033[0m] Execution failed in %s: %v\n", entryPath, err)
					return nil
				}
			} else {
				rt.L.SetGlobal("_CURRENT_FILE", lua.LString(entryPath))
				if err := rt.DoString(entryScript); err != nil {
					fmt.Printf("[\033[31mError\033[0m] Execution failed in stdin script: %v\n", err)
					return nil
				}
			}

			coordinator.FinishTask(entryPath)

			// Phase 3: Wait for async/dependent tasks
			coordinator.WaitWithoutClosing()

			// Phase 4: Schedule Lua finalizers
			rt.Schedule()

			// Phase 5: Capture results
			if err := coordinator.CaptureResults(rt, entryPath); err != nil {
				fmt.Printf("[\033[31mError\033[0m] Result capture failed for %s: %v\n", entryPath, err)
				return nil
			}

			// Phase 6: Detect cycles
			if err := coordinator.DetectCycles(); err != nil {
				fmt.Printf("[\033[31mError\033[0m] Dependency cycle detected: %v\n", err)
				return nil
			}

			// Phase 7: Commit changes
			if err := coordinator.Commit(); err != nil {
				fmt.Printf("[\033[31mError\033[0m] Commit failed: %v\n", err)
			} else {
				fmt.Printf("[\033[32mSuccess\033[0m] Build completed in %v\n", time.Since(start))
			}

			updateWatchList()

			return nil
		}

		// Initial run
		// Always watch the entry file before the first build so a startup failure
		// can still recover when the user fixes the file.
		updateWatchList()

		if err := execute(); err != nil {
			return err
		}

		fmt.Println("Watching for changes. Press 'q' to exit.")

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

		// Debouncing variables
		var debounceTimer *time.Timer
		var mu sync.Mutex

		// Rebuild queue
		rebuildChan := make(chan struct{}, 1)

		triggerRebuild := func() {
			mu.Lock()
			defer mu.Unlock()

			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(150*time.Millisecond, func() {
				select {
				case rebuildChan <- struct{}{}:
				default:
				}
			})
		}

		// Event loop
		for {
			select {
			case <-ctx.Done():
				return shutdown("\nReceived interrupt, shutting down...")
			case <-stopChan:
				return shutdown("Exiting watch mode...")
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}

				if ctx.Err() != nil {
					return shutdown("")
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

					triggerRebuild()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				fmt.Printf("Watcher error: %v\n", err)
			case <-rebuildChan:
				if ctx.Err() != nil {
					return shutdown("")
				}
				if err := execute(); err != nil {
					fmt.Printf("[\033[31mError\033[0m] Fatal error during rebuild: %v\n", err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
