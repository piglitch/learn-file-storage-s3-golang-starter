package main

import (
	"log"
	"os/exec"
)

func processVideoForFastRestart(filePath string) (string, error) {
	println("Enter processVideoForFastRestart")
	outputFilePath := filePath + ".processed.mp4"
	// inputPath := filePath + ".mp4"
	println("op: ", outputFilePath)
	println("ip: ", filePath)
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", outputFilePath)
	println("cmd executed")
	err := cmd.Run()
	if err != nil {
		log.Fatal("Failed to run the command: ", err)
		return "", err
	}
	println("Ran")
	return outputFilePath, nil
}