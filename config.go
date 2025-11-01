package main

import (
	"github.com/pelletier/go-toml/v2"
	"os"
)

type ShantyConfig struct {
	ServerUrl      string
	ServerUser     string
	ServerPassword string
}

func readConfig(conf *ShantyConfig) error {
	// Read config data.
	configData, err := os.ReadFile("config.toml")
	if err != nil {
		return err
	}

	// Extract data to object.
	err = toml.Unmarshal([]byte(configData), conf)
	return err
}
