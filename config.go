package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	FanLabels map[string]string `yaml:"fan_labels"`
}

func loadConfig() map[string]string {
	homeDir := ""
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		usr, err := user.Lookup(sudoUser)
		if err == nil {
			homeDir = usr.HomeDir
		}
	}
	if homeDir == "" {
		usr, err := user.Current()
		if err != nil {
			return map[string]string{}
		}
		homeDir = usr.HomeDir
	}
	configPath := filepath.Join(homeDir, ".config", "jayinsights", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return map[string]string{}
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return map[string]string{}
	}
	return cfg.FanLabels
}

func normalizeFanKey(s string) string {
	// e.g. "fan1_input", "Fan #1", "FAN 1", "FAN-1" → "Fan1"
	re := regexp.MustCompile(`(?i)fan[^0-9]*([0-9]+)`)
	match := re.FindStringSubmatch(s)
	if len(match) > 1 {
		return "Fan" + match[1]
	}
	return strings.TrimSpace(s)
}
