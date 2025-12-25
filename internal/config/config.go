package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	APIKey string
}

func Load() *Config {
	apiKey := os.Getenv("STRIPE_API_KEY")
	if apiKey == "" {
		cp, _ := os.UserHomeDir()
		path := filepath.Join(cp, ".lazystripe", "config")
		if data, err := os.ReadFile(path); err == nil {
			apiKey = strings.TrimSpace(string(data))
		} else {
			log.Fatal("Set STRIPE_API_KEY or run 'lazystripe init <key>'")
		}
	}
	return &Config{APIKey: apiKey}
}
