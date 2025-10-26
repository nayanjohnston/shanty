package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type ShantyConfig struct {
	ServerUrl      string
	ServerUser     string
	ServerPassword string
}

var songId = "CswcJyoHCNG9hsMuG8BMLm"
var songUrl = ""

func main() {
	var config ShantyConfig

	configData, err := os.ReadFile("config.toml")

	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal([]byte(configData), &config)

	songUrl = config.ServerUrl + "/rest/stream.view?u=" + config.ServerUser + "&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&id=" + songId

	fmt.Println(songUrl)
}
