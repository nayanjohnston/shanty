package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	albumArtWidth  = 20
	albumArtHeight = 6
)

func imageArray(imageId string) ([]string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	os.MkdirAll(cacheDir+"/shanty/art/", 0755)

	imageFile := cacheDir + "/shanty/art/" + imageId + ".jpg"
	textFile := cacheDir + "/shanty/art/" + imageId + ".txt"

	// File exists, just give them that.
	if _, err := os.Stat(textFile); err == nil {
		returnText, err := os.ReadFile(textFile)
		if err != nil {
			return nil, err
		}
		splitString := strings.Split(string(returnText), "\n")
		return splitString[:len(splitString)-1], nil
	}

	if _, err := os.Stat(imageFile); err != nil {
		imageUrl := config.ServerUrl +
			"/rest/getCoverArt?" +
			"u=" + config.ServerUser +
			"&p=" + config.ServerPassword +
			"&v=1.12.0" +
			"&c=shanty" +
			"&size=100" +
			"&id=" + imageId

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
		imageFile,
	).Output()

	file, err := os.Create(textFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, bytes.NewReader(chafaCommand))
	if err != nil {
		return nil, err
	}

	chafaOutput := string(chafaCommand)

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
