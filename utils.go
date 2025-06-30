package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"sync"
)

type cacheStatus string

const (
	local    cacheStatus = "local"
	cached   cacheStatus = "cached"
	notBuilt cacheStatus = "notBuilt"
)

type evalResult struct {
	Failed  bool
	Attr    string `json:"attr"`
	DrvPath string `json:"drvPath"`
	System  string `json:"system"`
	// lix doesn't report this, and will return `"isCached": true` regardless of reality
	CacheStatus cacheStatus `json:"cacheStatus"`
}

func startEvalJobs(cfg *Config, evalResultChan chan evalResult) {
	defer close(evalResultChan)

	evalCmd := exec.Command("nix-eval-jobs", cfg.evalArgs...)
	slog.Info("Calling nix-eval-jobs", "command", evalCmd.String())

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

		// eval fail
		if result.DrvPath == "" {
			result.Failed = true
		}

		slog.Debug("Received eval result", "raw", line, "result", result)

		// send all eval results to handler, let it decide whether/how to build
		evalResultChan <- result
	}

	err = evalCmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func queryOutputPaths(drvPath string) ([]string, error) {
	queryCmd := exec.Command("nix-store", "--query", drvPath)

	outputBytes, err := queryCmd.Output()
	if err != nil {
		slog.Error("Calling nix-store failed", "err", err)
		return nil, err
	}

	var results []string
	scanner := bufio.NewScanner(bytes.NewReader(outputBytes))
	for scanner.Scan() {
		results = append(results, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Unknown output from nix-store --query", "err", err)
		return nil, err
	}

	return results, nil
}

func evalResultHandler(cfg *Config, evalResultChan chan evalResult, builds *buildResults, wg *sync.WaitGroup) {
	defer wg.Done()

	for evalResult := range evalResultChan {
		slog.Info("Building eval result", "attr", evalResult.Attr)

		if evalResult.Failed {
			builds.addEvalFailed(evalResult.Attr)
			continue
		}

		// Successful eval result should be guaranteed to have a drvPath

		// cross build not supported yet
		if evalResult.System != cfg.system {
			builds.addFailed(evalResult.DrvPath)
			continue
		}

		if cfg.skipCached && evalResult.CacheStatus == cached {
			slog.Info("Skipping cached derivation", "attr", evalResult.Attr)
			builds.addSkipped(evalResult.Attr)
			continue
		}

		if cfg.forceSubstitute || evalResult.CacheStatus == notBuilt {
			buildCmd := exec.Command("nix-build", evalResult.DrvPath, "--out-link", path.Join(cfg.tmpDir, "builds", evalResult.Attr), "--keep-going", "--no-build-output")
			buildCmd.Stderr = os.Stderr
			err := buildCmd.Run()
			if err != nil {
				slog.Error("Build failed", "drv", evalResult.DrvPath, "error", err)
				builds.addFailed(evalResult.DrvPath)
				continue
			} else {
				slog.Info("Build succeeded", "attr", evalResult.Attr)
				builds.addSuccessful(evalResult.Attr)
			}
		} else {
			slog.Info("Derivation already built", "attr", evalResult.Attr)
			builds.addSkipped(evalResult.Attr)
		}

		if cfg.atticCache != "" {
			outputPaths, err := queryOutputPaths(evalResult.DrvPath)
			if err != nil {
				slog.Error("Unable to qeury derivation output", "drv", evalResult.DrvPath)
				continue
			}
			for _, path := range outputPaths {
				slog.Info("Pushing output to attic", "output", path)
				atticCmd := exec.Command("attic", "push", cfg.atticCache, path)
				atticCmd.Stderr = os.Stderr
				err := atticCmd.Run()
				if err != nil {
					slog.Error("Attic push failed", "error", err)
				}
			}
		}
	}
}
