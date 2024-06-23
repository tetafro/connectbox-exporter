package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents application configuration.
type Config struct {
	ListenAddr string        `yaml:"listen_addr"`
	Timeout    time.Duration `yaml:"timeout"`
	Targets    []Target      `yaml:"targets"`
}

// Target is a single ConnectBox device.
type Target struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// ReadConfig returns configuration populated from the config file.
func ReadConfig(file string) (Config, error) {
	data, err := os.ReadFile(file) //nolint:gosec
	if err != nil {
		return Config{}, fmt.Errorf("read file: %w", err)
	}
	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return Config{}, fmt.Errorf("unmarshal file: %w", err)
	}

	// Set defaults
	if conf.ListenAddr == "" {
		conf.ListenAddr = "0.0.0.0:9119"
	}
	if conf.Timeout == 0 {
		conf.Timeout = 30 * time.Second
	}
	for i := range conf.Targets {
		if conf.Targets[i].Addr == "" {
			return Config{}, fmt.Errorf("found target with empty address")
		}
		if conf.Targets[i].Username == "" {
			conf.Targets[i].Username = "NULL"
		}
		if conf.Targets[i].Password == "" {
			return Config{}, fmt.Errorf("found target with empty password")
		}
	}

	return conf, nil
}
