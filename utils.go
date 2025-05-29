package main

import (
	"bufio"
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
	Attr    string `json:"attr"`
	DrvPath string `json:"drvPath"`
	System  string `json:"system"`
	// lix doesn't report this, and will return `"isCached": true` regardless of reality
	CacheStatus cacheStatus `json:"cacheStatus"`
	// in Lix, this field is null and another `inputDrvs` is added
	// to avoid this, we simply consider the path with `.drv` trimmed as output
	// Outputs     map[string]string `json:"outputs"`
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

		slog.Debug("Handling eval result", "raw", line, "result", result)

		if result.System == cfg.system {
			evalResultChan <- result
		}
	}

	err = evalCmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func evalResultHandler(cfg *Config, evalResultChan chan evalResult, builds *buildResults, wg *sync.WaitGroup) {
	defer wg.Done()

	for evalResult := range evalResultChan {
		slog.Info("Handling eval result", "attr", evalResult.Attr)

		if cfg.skipCached && evalResult.CacheStatus == cached {
			slog.Info("Skipping cached derivation", "attr", evalResult.Attr)
			builds.addSkipped(evalResult.Attr)
			continue
		}

		if cfg.forceSubstitute || evalResult.CacheStatus == notBuilt {
			buildCmd := exec.Command("nix-build", evalResult.DrvPath, "--out-link", path.Join(cfg.tmpDir, "builds", evalResult.Attr))
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
			// for _, out := range evalResult.Outputs {
			// 	slog.Info("Pushing output to attic", "output", out)
			// 	atticCmd := exec.Command("attic", "push", cfg.atticCache, out)
			// 	atticCmd.Stderr = os.Stderr
			// 	err := atticCmd.Run()
			// 	if err != nil {
			// 		slog.Error("Attic push failed", "error", err)
			// 	}
			// }
			buildResult := evalResult.DrvPath[:len(evalResult.DrvPath)-4] // trim `.drv` suffix
			slog.Info("Pushing output to attic", "output", buildResult)
			atticCmd := exec.Command("attic", "push", cfg.atticCache, buildResult)
			atticCmd.Stderr = os.Stderr
			err := atticCmd.Run()
			if err != nil {
				slog.Error("Attic push failed", "error", err)
			}
		}
	}
}
