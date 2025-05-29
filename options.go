package main

import (
	"flag"
	"log/slog"
)

type Config struct {
	targetAttr string
	evalArgs   []string
}

func defaultConfig() *Config {
	return &Config{
		evalArgs: []string{},
	}
}

func loadConfig() (*Config, error) {
	config := defaultConfig()

	flag.StringVar(&config.targetAttr, "f", ".#nixosConfigurations", "Target flakes attribute to build")

	flag.Func("evalArgs", "Arguments passed directly to nix-eval-jobs", func(value string) error {
		config.evalArgs = append(config.evalArgs, value)
		return nil
	})

	flag.Parse()

	slog.Info("Loading configuration", "config", config)

	return config, nil
}
