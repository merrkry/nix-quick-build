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

	var wg sync.WaitGroup
	evals := make(chan evalResult, 1024)
	wg.Add(config.numWorkers)
	for i := 0; i < config.numWorkers; i++ {
		go evalResultHandler(config, evals, &builds, &wg)
	}
	startEvalJobs(config, evals)
	wg.Wait()

	builds.printResults()

	if len(builds.failedBuilds) > 0 || len(builds.evalFailedBuilds) > 0 {
		slog.Error("Some builds failed, check the output above for details.")

		// TODO: handle this more gracefully
		os.RemoveAll(config.tmpDir)

		os.Exit(1)
	} else {
		slog.Info("All builds completed successfully.")
	}
}
