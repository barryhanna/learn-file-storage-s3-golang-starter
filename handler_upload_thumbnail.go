package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", err)
		return
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", nil)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable get image meta data", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Must be owner of requested video", err)
		return
	}
	// videoThumbnail := thumbnail{data: imageData, mediaType: mediaType}
	// videoThumbnails[videoID] = videoThumbnail

	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	urlEncodedVideoFilename := base64.RawURLEncoding.EncodeToString(randomBytes)
	fileExtension := strings.Split(mediaType, "/")[1]
	path := filepath.Join(cfg.assetsRoot, urlEncodedVideoFilename+"."+fileExtension)
	formattedThumbnailUrl := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, urlEncodedVideoFilename, fileExtension)
	// formattedThumbnailUrl := fmt.Sprintf("data:%s;base64,%s", mediaType, base64.StdEncoding.EncodeToString(imageData))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Problem generating random thumbnail name", err)
		return
	}
	outFile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Problem saving thumnail file to filesystem", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Problem saving thumnail file to filesystem", err)
		return
	}
	video.ThumbnailURL = &formattedThumbnailUrl
	cfg.db.UpdateVideo(video)
	respondWithJSON(w, http.StatusOK, database.Video(video))
}
