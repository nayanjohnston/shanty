package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type ShantyConfig struct {
	ServerUrl      string
	ServerUser     string
	ServerPassword string
	ShouldScrobble bool
}

func readConfig(conf *ShantyConfig) error {
	// Set defaults
	conf.ShouldScrobble = true

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Read config data.
	configData, err := os.ReadFile(homeDir + "/.config/shanty/config.toml")
	if err != nil {
		return err
	}

	// Extract data to object.
	err = toml.Unmarshal([]byte(configData), conf)

	return err
}
