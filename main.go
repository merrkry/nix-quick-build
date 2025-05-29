package main

import (
	"sync"
)

func main() {
	config, _ := loadConfig()

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
