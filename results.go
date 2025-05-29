package main

import (
	"sync"
)

type buildResults struct {
	mu sync.Mutex
	// for skipped/successful builds, we only store attr name for simplicity
	// for failed builds, we store full drv path for debugging purposes
	skippedBuilds    []string
	successfulBuilds []string
	failedBuilds     []string
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
