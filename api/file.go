// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/clivern/tut/db"
	"github.com/clivern/tut/middleware"
	"github.com/clivern/tut/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// getStoragePath returns the base storage path
func getStoragePath() string {
	basePath := viper.GetString("app.storage.path")
	if basePath == "" {
		basePath = "./storage"
	}
	return basePath
}

// UploadFile handles file upload to a bucket
func UploadFile(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Upload file endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	// Verify bucket exists and user has access
	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByID(bucketID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve bucket",
		})
		return
	}

	if bucket == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Bucket not found",
		})
		return
	}

	if bucket.UserID != user.ID && !bucket.IsPublic {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied",
		})
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(100 << 20) // 100 MB max
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Failed to parse form",
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "No file provided",
		})
		return
	}
	defer file.Close()

	fileName := header.Filename
	if fileName == "" {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid file name",
		})
		return
	}

	// Check if file already exists
	fileRepo := db.NewFileRepository(db.GetDB())
	existingFile, err := fileRepo.GetByName(bucketID, fileName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing file")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to upload file",
		})
		return
	}

	// Create storage directory structure: storage/user_id/bucket_id/
	storageBase := getStoragePath()
	storageDir := filepath.Join(storageBase, fmt.Sprintf("%d", user.ID), fmt.Sprintf("%d", bucketID))
	if err := service.EnsureDir(storageDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create storage directory")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create storage directory",
		})
		return
	}

	// Create file path
	filePath := filepath.Join(storageDir, fileName)

	// Create file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create file")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to save file",
		})
		return
	}
	defer dst.Close()

	// Copy file content and calculate hash
	hash := md5.New()
	multiWriter := io.MultiWriter(dst, hash)
	size, err := io.Copy(multiWriter, file)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save file")
		os.Remove(filePath)
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to save file",
		})
		return
	}

	etag := fmt.Sprintf("%x", hash.Sum(nil))
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Save file metadata to database
	dbFile := &db.File{
		BucketID:    bucketID,
		Name:        fileName,
		Path:        filePath,
		ContentType: contentType,
		Size:        size,
		ETag:        etag,
		UserID:      user.ID,
	}

	if existingFile != nil {
		// Update existing file
		dbFile.ID = existingFile.ID
		if err := fileRepo.Update(dbFile); err != nil {
			log.Error().Err(err).Msg("Failed to update file")
			os.Remove(filePath)
			service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": "Failed to save file metadata",
			})
			return
		}
		log.Info().
			Int64("file_id", dbFile.ID).
			Str("file_name", fileName).
			Int64("bucket_id", bucketID).
			Msg("File updated successfully")
	} else {
		// Create new file
		if err := fileRepo.Create(dbFile); err != nil {
			log.Error().Err(err).Msg("Failed to create file")
			os.Remove(filePath)
			service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": "Failed to save file metadata",
			})
			return
		}
		log.Info().
			Int64("file_id", dbFile.ID).
			Str("file_name", fileName).
			Int64("bucket_id", bucketID).
			Msg("File uploaded successfully")
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":           dbFile.ID,
		"name":         dbFile.Name,
		"bucket_id":    dbFile.BucketID,
		"content_type": dbFile.ContentType,
		"size":         dbFile.Size,
		"etag":         dbFile.ETag,
		"created_at":   dbFile.CreatedAt,
		"updated_at":   dbFile.UpdatedAt,
	})
}

// ListFiles lists all files in a bucket
func ListFiles(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List files endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	// Verify bucket exists and user has access
	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByID(bucketID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve bucket",
		})
		return
	}

	if bucket == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Bucket not found",
		})
		return
	}

	if bucket.UserID != user.ID && !bucket.IsPublic {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied",
		})
		return
	}

	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Parse prefix filter
	prefix := r.URL.Query().Get("prefix")

	fileRepo := db.NewFileRepository(db.GetDB())
	var files []*db.File
	var total int64

	if prefix != "" {
		files, err = fileRepo.ListByPrefix(bucketID, prefix, limit, offset)
	} else {
		files, err = fileRepo.List(bucketID, limit, offset)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to list files")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list files",
		})
		return
	}

	total, err = fileRepo.Count(bucketID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count files")
	}

	result := make([]map[string]interface{}, len(files))
	for i, file := range files {
		result[i] = map[string]interface{}{
			"id":           file.ID,
			"name":         file.Name,
			"bucket_id":    file.BucketID,
			"content_type": file.ContentType,
			"size":         file.Size,
			"etag":         file.ETag,
			"created_at":   file.CreatedAt,
			"updated_at":   file.UpdatedAt,
		}
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"files":  result,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetFile retrieves a file by ID
func GetFile(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get file endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	fileIDStr := chi.URLParam(r, "fileId")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid file ID",
		})
		return
	}

	// Verify bucket exists and user has access
	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByID(bucketID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve bucket",
		})
		return
	}

	if bucket == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Bucket not found",
		})
		return
	}

	if bucket.UserID != user.ID && !bucket.IsPublic {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied",
		})
		return
	}

	fileRepo := db.NewFileRepository(db.GetDB())
	file, err := fileRepo.GetByID(fileID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get file")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve file",
		})
		return
	}

	if file == nil || file.BucketID != bucketID {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "File not found",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":           file.ID,
		"name":         file.Name,
		"bucket_id":    file.BucketID,
		"content_type": file.ContentType,
		"size":         file.Size,
		"etag":         file.ETag,
		"created_at":   file.CreatedAt,
		"updated_at":   file.UpdatedAt,
	})
}

// DownloadFile downloads a file
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Download file endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	fileIDStr := chi.URLParam(r, "fileId")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid file ID",
		})
		return
	}

	// Verify bucket exists and user has access
	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByID(bucketID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve bucket",
		})
		return
	}

	if bucket == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Bucket not found",
		})
		return
	}

	if bucket.UserID != user.ID && !bucket.IsPublic {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied",
		})
		return
	}

	fileRepo := db.NewFileRepository(db.GetDB())
	file, err := fileRepo.GetByID(fileID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get file")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve file",
		})
		return
	}

	if file == nil || file.BucketID != bucketID {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "File not found",
		})
		return
	}

	// Check if file exists on disk
	if !service.FileExists(file.Path) {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "File not found on disk",
		})
		return
	}

	// Open file
	fileData, err := os.Open(file.Path)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open file")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to read file",
		})
		return
	}
	defer fileData.Close()

	// Set headers
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	w.Header().Set("ETag", file.ETag)

	// Copy file to response
	_, err = io.Copy(w, fileData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send file")
		return
	}

	log.Info().
		Int64("file_id", fileID).
		Int64("bucket_id", bucketID).
		Msg("File downloaded successfully")
}

// DeleteFile deletes a file
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Delete file endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	fileIDStr := chi.URLParam(r, "fileId")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid file ID",
		})
		return
	}

	// Verify bucket exists and user has access
	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByID(bucketID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve bucket",
		})
		return
	}

	if bucket == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "Bucket not found",
		})
		return
	}

	if bucket.UserID != user.ID {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied",
		})
		return
	}

	fileRepo := db.NewFileRepository(db.GetDB())
	file, err := fileRepo.GetByID(fileID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get file")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to retrieve file",
		})
		return
	}

	if file == nil || file.BucketID != bucketID {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "File not found",
		})
		return
	}

	// Delete file from disk
	if service.FileExists(file.Path) {
		if err := os.Remove(file.Path); err != nil {
			log.Error().Err(err).Msg("Failed to delete file from disk")
		}
	}

	// Delete file from database
	if err := fileRepo.Delete(fileID); err != nil {
		log.Error().Err(err).Msg("Failed to delete file")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete file",
		})
		return
	}

	log.Info().
		Int64("file_id", fileID).
		Int64("bucket_id", bucketID).
		Int64("user_id", user.ID).
		Msg("File deleted successfully")

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "File deleted successfully",
	})
}
