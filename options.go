package main

import (
	"bytes"
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"path"
)

type Config struct {
	targetAttr      string
	skipCached      bool
	evalArgs        []string
	tmpDir          string
	logLevel        slog.Level
	atticCache      string
	system          string
	forceSubstitute bool
	// TODO: worker, handler limits
}

func defaultConfig() (*Config, error) {
	tmpDir, err := os.MkdirTemp("", "nix-quick-build")
	if err != nil {
		return nil, err
	}

	var arch bytes.Buffer
	archCmd := exec.Command("nix", "eval", "--raw", "--impure", "--expr", "builtins.currentSystem")
	archCmd.Stdout = &arch
	err = archCmd.Run()
	if err != nil {
		slog.Error("Unable to get system architecture infomation.")
		return nil, err
	}

	return &Config{
		targetAttr:      ".#nixosConfigurations",
		skipCached:      false,
		evalArgs:        []string{},
		tmpDir:          tmpDir,
		logLevel:        slog.LevelInfo,
		atticCache:      "",
		system:          arch.String(),
		forceSubstitute: false,
	}, nil
}

func loadConfig() (*Config, error) {
	defaultCfg, err := defaultConfig()
	if err != nil {
		return nil, err
	}
	cfg := defaultCfg

	flag.StringVar(&cfg.targetAttr, "flake", defaultCfg.targetAttr, "Target flakes attribute to build")

	flag.BoolVar(&cfg.skipCached, "skip-cached", defaultCfg.skipCached, "Derivation already cached by any substituter will not be built/uploaded. Disabled by default.")

	flag.Func("args", "Extra arguments passed directly to nix-eval-jobs", func(value string) error {
		cfg.evalArgs = append(cfg.evalArgs, value)
		return nil
	})

	flag.Func("log-level", "Log level as defined in log/slog", func(value string) error {
		return cfg.logLevel.UnmarshalText([]byte(value))
	})

	flag.StringVar(&cfg.atticCache, "attic-cache", defaultCfg.atticCache, "Attic cache name")

	flag.BoolVar(&cfg.forceSubstitute, "force-substitute", defaultCfg.forceSubstitute, "Substitute locally regardless of reported cache status. Useful with lix.")

	flag.Parse()

	cfg.evalArgs = append(cfg.evalArgs, "--flake", cfg.targetAttr, "--force-recurse", "--gc-roots-dir", path.Join(cfg.tmpDir, "evals"), "--check-cache-status")

	slog.SetLogLoggerLevel(cfg.logLevel)

	slog.Info("Loading configuration", "config", cfg)

	return cfg, nil
}
