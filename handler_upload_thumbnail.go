package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	// "mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
	// "golang.org/x/tools/go/analysis/passes/defers"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	maxMemory := 10 << 20
	err = r.ParseMultipartForm(int64(maxMemory))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse multipart form", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user ", userID)

	// TODO: implement the upload here

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not read rand key", err)
		return
	}
	fileNameStr := base64.RawURLEncoding.EncodeToString(key)

	fileData, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer fileData.Close()
	
	vid, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video from db", err)
		return
	} 
	if vid.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", err)
		return
	}

	mediaType := fileHeader.Header.Get("Content-Type")

	if mediaType[:5] != "image" {
		respondWithError(w, http.StatusBadRequest, "Not an image", err)
		return
	}

	FILE_PATH := filepath.Join(cfg.assetsRoot, fileNameStr) + "." + mediaType[6:] 

	file, err := os.Create(FILE_PATH)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to create a new file", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, fileData)
	
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to copy data to a new file", err)
		return
	}
	
	thumbnailUrl := "http://localhost:" + cfg.port + "/assets/" + fileNameStr + "." + mediaType[6:]

	vidParams := database.CreateVideoParams{
		Title: vid.Title,
		Description: vid.Description,
		UserID: vid.UserID,
	}
	newVideo := database.Video{
		ID: vid.ID,
		ThumbnailURL: &thumbnailUrl,
		UpdatedAt: time.Now(),
		CreateVideoParams: vidParams,
	}
	err = cfg.db.UpdateVideo(newVideo)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to update video (db)", err)
		return
	}
	respondWithJSON(w, http.StatusOK, newVideo)
}
