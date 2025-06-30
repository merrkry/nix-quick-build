package main

import (
	"fmt"
	"sync"
)

type buildResults struct {
	mu sync.Mutex
	// for skipped/successful builds, we only store attr name for simplicity
	// for failed builds, we store full drv path for debugging purposes
	skippedBuilds    []string
	successfulBuilds []string
	failedBuilds     []string
	evalFailedBuilds []string
}

func (br *buildResults) addSkipped(attr string) {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.skippedBuilds = append(br.skippedBuilds, attr)
}

func (br *buildResults) addSuccessful(attr string) {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.successfulBuilds = append(br.successfulBuilds, attr)
}

func (br *buildResults) addFailed(drvPath string) {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.failedBuilds = append(br.failedBuilds, drvPath)
}

func (br *buildResults) addEvalFailed(attr string) {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.evalFailedBuilds = append(br.evalFailedBuilds, attr)
}

func (br *buildResults) printResults() {
	br.mu.Lock()
	defer br.mu.Unlock()

	fmt.Println("Skipped builds:")
	for _, attr := range br.skippedBuilds {
		fmt.Println("-", attr)
	}

	fmt.Println("Successful builds:")
	for _, attr := range br.successfulBuilds {
		fmt.Println("-", attr)
	}

	fmt.Println("Failed builds:")
	for _, drvPath := range br.failedBuilds {
		fmt.Println("-", drvPath)
	}

	fmt.Println("Eval failed builds:")
	for _, attr := range br.evalFailedBuilds {
		fmt.Println("-", attr)
	}
}
