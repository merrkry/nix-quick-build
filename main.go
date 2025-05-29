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

	numWorkers := 4
	var wg sync.WaitGroup
	resultChan := make(chan evalResult, 128)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go evalResultHandler(resultChan, &wg)
	}
	startEvalJobs(config, resultChan)
	wg.Wait()
}
