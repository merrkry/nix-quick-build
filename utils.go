package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"sync"
)

type cacheStatus string

const (
	local    cacheStatus = "local"
	cached   cacheStatus = "cached"
	notBuilt cacheStatus = "notBuilt"
)

type evalResult struct {
	Attr        string      `json:"attr"`
	DrvPath     string      `json:"drvPath"`
	System      string      `json:"system"`
	CacheStatus cacheStatus `json:"cacheStatus"`
}

func startEvalJobs(cfg *Config, evalResultChan chan evalResult) {
	defer close(evalResultChan)

	evalCmd := exec.Command("nix-eval-jobs", cfg.evalArgs...)

	// evalCmd.Stderr = io.Discard
	evalCmd.Stderr = os.Stderr
	stdout, err := evalCmd.StdoutPipe()
	if err != nil {
		log.Fatal("Unable to create stdout pipe", err)
	}

	err = evalCmd.Start()
	if err != nil {
		log.Fatal("Unable to start nix-eval-jobs", err)
	}

	reader := bufio.NewReader(stdout)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal("Invalid eval result", err, line)
			}
		}

		var result evalResult
		err = json.Unmarshal([]byte(line), &result)
		if err != nil {
			continue
		}
		
		slog.Debug("Handling eval result", "raw", line, "result", result)

		evalResultChan <- result
	}

	err = evalCmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func evalResultHandler(evalResultChan chan evalResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for evalResult := range evalResultChan {
		switch evalResult.CacheStatus {
		case local:
			slog.Info(fmt.Sprintf("Local derivation: %s", evalResult.Attr))
		case cached:
			slog.Info(fmt.Sprintf("Cached derivation: %s", evalResult.Attr))
		case notBuilt:
			slog.Info(fmt.Sprintf("Not built derivation: %s", evalResult.Attr))
		}
	}
}
