package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
)

type Dimensions struct {
	Streams []struct{
		DisplayAspectRatio string `json:"display_aspect_ratio"`
	} `json:"streams"`	
}

func getVideoAspectRatio(filepath string) (string, error) {

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)
	b := &bytes.Buffer{}
	cmd.Stdout = b
	cmd.Stderr = b
	err := cmd.Run()
	if err != nil {
		log.Fatal("Failed to run the command: ", filepath)
		return "", err
	}
	
	dimensions := Dimensions{}
	err = json.Unmarshal(b.Bytes(), &dimensions)
	if err != nil {
		log.Fatal("Failed to unmarshal b.bytes")
		return "", err
	}
	return b.String(), nil

}
