package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
		return errors.New("shanty: No config file found. Create one in \"~/.config/shanty/config.toml\"")
	}

	// Read config data.
	configData, err := os.ReadFile(homeDir + "/.config/shanty/config.toml")
	if err != nil {
		return errors.New("shanty: Config data is invalid. Make sure all values are correct.")
	}

	// Extract data to object.
	err = toml.Unmarshal([]byte(configData), conf)

	// Validate values...
	serverUrl := conf.ServerUrl +
		"/rest/ping.view?" +
		"u=" + conf.ServerUser +
		"&p=" + conf.ServerPassword +
		"&v=1.12.0" +
		"&c=shanty" +
		"&f=json"

	result, err := http.Get(serverUrl)
	if err != nil {
		return errors.New("shanty: URL cannot be accessed. (Is it valid?)")
	}

	resultBody, _ := io.ReadAll(result.Body)

	var list any
	json.Unmarshal([]byte(resultBody), &list)

	respSubsonic := list.(map[string]any)["subsonic-response"]
	status := respSubsonic.(map[string]any)["status"].(string)

	if status != "ok" {
		urlError := respSubsonic.(map[string]any)["error"]
		message := urlError.(map[string]any)["message"].(string)
		return errors.New("config: " + message)
	}

	return err
}
