package main

import (
	"bytes"
	"os/exec"
	"strings"
)

func imageArray(imageId string) []string {
	var b bytes.Buffer

	c1 := exec.Command("curl", config.ServerUrl+"/rest/getCoverArt?u="+config.ServerUser+"&p="+config.ServerPassword+"&v=1.12.0&c=shanty&id="+imageId)

	c2 := exec.Command("chafa", "-s", "20x20", "-f", "symbols")

	c2.Stdin, _ = c1.StdoutPipe()
	c2.Stdout = &b
	_ = c2.Start()
	_ = c1.Run()
	_ = c2.Wait()

	return strings.Split(b.String(), "\n")
}
