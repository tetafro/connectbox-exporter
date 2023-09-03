package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents application configuration.
type Config struct {
	ListenAddr string   `yaml:"listen_addr"`
	Targets    []Target `yaml:"targets"`
}

// Target is a single ConnectBox device.
type Target struct {
	Addr     string `yaml:"ziggo_addr"`
	Username string `yaml:"ziggo_username"`
	Password string `yaml:"ziggo_password"`
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
	return conf, nil
}
