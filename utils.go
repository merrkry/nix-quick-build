package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type evalJobArgs struct {
	attr string
}

func (args evalJobArgs) createEvalCmd() *exec.Cmd {
	return exec.Command("nix-eval-jobs", "--flake", args.attr, "--gc-roots-dir", "gcroots", "--force-recurse")
}

type evalResult struct {
	Attr    string `json:"attr"`
	DrvPath string `json:"drvPath"`
	System  string `json:"system"`
}

func startEvalJobs(args evalJobArgs, evalResultChan chan evalResult) {
	defer close(evalResultChan)

	evalCmd := args.createEvalCmd()

	// evalCmd.Stderr = io.Discard
	evalCmd.Stderr = os.Stderr
	stdout, err := evalCmd.StdoutPipe()
	if err != nil {
		log.Fatal("Unable to create stdout pipe", err)
	}

	err = evalCmd.Start()
	if err != nil {
		log.Fatal("Unable to start eval command", err)
	}

	reader := bufio.NewReader(stdout)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal("Invalid eval result", err)
			}
		}

		line = strings.TrimSpace(line)

		var result evalResult
		err = json.Unmarshal([]byte(line), &result)
		if err != nil {
			continue
		}

		evalResultChan <- result
	}

	fmt.Println("Waiting nix-eval-jobs to exit.")

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
