package main

import (
	"log/slog"
	"os"
	"sync"
)

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
	evals := make(chan evalResult, 128)
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go evalResultHandler(config, evals, &builds, &wg)
	}
	startEvalJobs(config, evals)
	wg.Wait()

	slog.Info("Build results:", "skipped", builds.skippedBuilds, "successful", builds.successfulBuilds, "failed", builds.failedBuilds)
}
