package main

import (
	"log/slog"
	"os"
	"os/exec"
	"sync"
)

func init() {
	var requiredExecutables = []string{"nix", "nix-eval-jobs", "attic"}

	for _, executable := range requiredExecutables {
		if _, err := exec.LookPath(executable); err != nil {
			slog.Error("Required executable not found", "executable", executable)
			panic(err)
		}
	}
}

func main() {
	config, err := loadConfig()
	defer os.RemoveAll(config.tmpDir)
	if err != nil {
		slog.Error("initialization failed.")
		panic(err)
	}

	builds := buildResults{
		mu:               sync.Mutex{},
		skippedBuilds:    []string{},
		successfulBuilds: []string{},
		failedBuilds:     []string{},
	}

	numWorkers := 4
	var wg sync.WaitGroup
	evals := make(chan evalResult, 128) // TODO: dynamic limit
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go evalResultHandler(config, evals, &builds, &wg)
	}
	startEvalJobs(config, evals)
	wg.Wait()

	builds.printResults()
}
