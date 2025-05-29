package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func createEvalCmd(cfg *Config) *exec.Cmd {
	return exec.Command("nix-eval-jobs", "--flake", cfg.targetAttr, "--gc-roots-dir", "gcroots", "--force-recurse", "--check-cache-status")
}

func startEvalJobs(cfg *Config, evalResultChan chan evalResult) {
	defer close(evalResultChan)

	evalCmd := createEvalCmd(cfg)

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
		// placeholder
		fmt.Println(evalResult)
	}
}
