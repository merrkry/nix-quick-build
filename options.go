package main

import (
	"flag"
	"io/ioutil"
	"log/slog"
	"path"
)

type Config struct {
	targetAttr string
	skipCached bool
	evalArgs   []string
	tmpDir     string
	logLevel   slog.Level
	// TODO: worker, handler limits
}

func defaultConfig() (*Config, error) {
	tmpDir, err := ioutil.TempDir("", "nix-quick-build")
	if err != nil {
		return nil, err
	}
	return &Config{
		targetAttr: ".#nixosConfigurations",
		skipCached: false,
		evalArgs:   []string{},
		tmpDir:     tmpDir,
		logLevel:   slog.LevelInfo,
	}, nil
}

func loadConfig() (*Config, error) {
	defaultCfg, err := defaultConfig()
	if err != nil {
		return nil, err
	}
	cfg := defaultCfg

	flag.StringVar(&cfg.targetAttr, "flake", defaultCfg.targetAttr, "Target flakes attribute to build")

	flag.BoolVar(&cfg.skipCached, "skip-cached", defaultCfg.skipCached, "Skip building or uploading drv already cached by substituters. Disabled by default.")

	flag.Func("args", "Extra arguments passed directly to nix-eval-jobs", func(value string) error {
		cfg.evalArgs = append(cfg.evalArgs, value)
		return nil
	})

	flag.Func("log-level", "Log level as defined in log/slog", func(value string) error {
		return cfg.logLevel.UnmarshalText([]byte(value))
	})

	flag.Parse()

	cfg.evalArgs = append(cfg.evalArgs, "--flake", cfg.targetAttr, "--force-recurse", "--gc-roots-dir", path.Join(cfg.tmpDir, "gcroots"), "--check-cache-status")

	slog.SetLogLoggerLevel(cfg.logLevel)

	slog.Info("Loading configuration", "config", cfg)

	return cfg, nil
}
