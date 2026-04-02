// Package main is the entry point for the cogeni CLI tool.
// It initializes the command-line interface and delegates execution to the cmd package.
package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/lucasfabre/cogeni/src/cmd"
)

func main() {
	// Enable CPU profiling if CPUPROFILE environment variable is set
	if cpuProfile := os.Getenv("CPUPROFILE"); cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
