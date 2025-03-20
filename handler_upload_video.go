package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}
	maxMemory := 10 << 30
	http.MaxBytesReader(w, r.Body, int64(maxMemory))

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not validate JWT", err)
		return
	}
	fmt.Println("uploading video", videoID, "by user ", userID)

	vidData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video from db", err)
		return
	}
	if vidData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not read rand key", err)
		return
	}
	fileNameStr := base64.RawURLEncoding.EncodeToString(key) + ".mp4"

	fileData, fileHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer fileData.Close()

	mediaType := fileHeader.Header.Get("Content-Type")

	if mediaType[:5] != "video" {
		respondWithError(w, http.StatusBadRequest, "Not a video", err)
		return
	}

	println("filename: ", fileNameStr)
	file, err := os.CreateTemp("", fileNameStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not create temp file", err)
		return
	}
	defer os.Remove(file.Name())
	defer file.Close()
	
	_, err = io.Copy(file, fileData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to copy data to a new file", err)
		return
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to set offset", err)
		return
	}

	aspectRatio, err := getVideoAspectRatio(file.Name())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get the aspect ratio", err)
		return
	}
	
	dimensions := Dimensions{}
	err = json.Unmarshal([]byte(aspectRatio), &dimensions)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to unmarshal Dimension{}", err)
		return
	}

	aspect_ratio := dimensions.Streams[0].DisplayAspectRatio

	orientation := "other" 
	if aspect_ratio == "16:9" {
		orientation = "landscape"
	}
	if aspect_ratio == "9:16" {
		orientation = "portrait"
	}

	fileNameStr = orientation + "/" + fileNameStr

	cfg.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         &fileNameStr,
		Body:        file,
		ContentType: &mediaType,
	})

	videoUrl := "https://" + cfg.s3Bucket + ".s3." + cfg.s3Region + ".amazonaws.com/" + fileNameStr

	vidParams := database.CreateVideoParams{
		Title:       vidData.Title,
		Description: vidData.Description,
		UserID:      vidData.UserID,
	}

	newVideo := database.Video{
		ID:                vidData.ID,
		CreatedAt:         vidData.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      vidData.ThumbnailURL,
		VideoURL:          &videoUrl,
		CreateVideoParams: vidParams,
	}
	err = cfg.db.UpdateVideo(newVideo)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to Update the video (db)", err)
		return
	}
	respondWithJSON(w, http.StatusOK, newVideo)
}
