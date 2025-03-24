package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	videoUrl := video.VideoURL
	if videoUrl == nil {
		println("No video uploaded")
		return database.Video{}, nil
	}
	fmt.Println("video url: ", *videoUrl)
	urlSlice := strings.Split(*videoUrl, ",")
	if len(urlSlice) < 2 {
		println("No video uploaded")
		return database.Video{}, nil
	}
	bucket := urlSlice[0]
	key := urlSlice[1]
	sevenDays := 7 * 24 * time.Hour
	preSignedUrl, err := generatePresignedURL(cfg.s3Client, bucket, key, sevenDays)
	if err != nil {
		log.Fatal("Error generating presigned url")
		return database.Video{}, err
	}
	video.VideoURL = &preSignedUrl
	return video, nil
}