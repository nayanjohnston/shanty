package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	albumArtWidth  = 16
	albumArtHeight = 5
)

func imageArray(imageId string) ([]string, error) {
	imageFile := "./.tmp/" + imageId + ".jpg"

	if _, err := os.Stat(imageFile); err != nil {
		imageUrl := config.ServerUrl + "/rest/getCoverArt?u=" + config.ServerUser +
			"&p=" + config.ServerPassword + "&v=1.12.0&c=shanty&size=100&id=" + imageId

		imageResponse, err := http.Get(imageUrl)
		if err != nil {
			return nil, err
		}
		defer imageResponse.Body.Close()

		file, err := os.Create(imageFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = io.Copy(file, imageResponse.Body)
		if err != nil {
			return nil, err
		}
	}

	chafaCommand, err := exec.Command(
		"chafa",
		"-s",
		fmt.Sprintf("%vx%v", albumArtWidth, albumArtHeight),
		"-f",
		"symbols",
		"--align",
		"hcenter",
		"--view-size",
		fmt.Sprintf("%vx%v", albumArtWidth, albumArtHeight),
		imageFile,
	).Output()

	chafaOutput := string(chafaCommand)

	if err != nil {
		return nil, err
	}

	splitString := strings.Split(chafaOutput, "\n")
	return splitString[:len(splitString)-1], nil
}

func drawImage(image []string) string {
	s := ""

	for index, row := range image {
		s += row
		if index != len(image)-1 {
			s += "\n"
		}
	}

	return s
}
