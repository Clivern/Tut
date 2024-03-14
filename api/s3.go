// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/clivern/tut/db"
	"github.com/clivern/tut/middleware"
	"github.com/clivern/tut/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// S3 XML response structures
type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Buckets Buckets  `xml:"Buckets"`
	Owner   Owner    `xml:"Owner"`
}

type Buckets struct {
	Bucket []BucketInfo `xml:"Bucket"`
}

type BucketInfo struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type ListBucketResult struct {
	XMLName        xml.Name       `xml:"ListBucketResult"`
	Name           string         `xml:"Name"`
	Prefix         string         `xml:"Prefix"`
	Marker         string         `xml:"Marker"`
	MaxKeys        int            `xml:"MaxKeys"`
	IsTruncated    bool           `xml:"IsTruncated"`
	Contents       []Content      `xml:"Contents"`
	CommonPrefixes []CommonPrefix `xml:"CommonPrefixes"`
}

type Content struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
	Owner        Owner  `xml:"Owner"`
}

type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// S3ListBuckets handles S3-compatible bucket listing
// GET /?list-type=2
func S3ListBuckets(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("S3 list buckets endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketRepo := db.NewBucketRepository(db.GetDB())
	buckets, err := bucketRepo.List(user.ID, 1000, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list buckets")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bucketList := make([]BucketInfo, len(buckets))
	for i, bucket := range buckets {
		bucketList[i] = BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreatedAt.Format(time.RFC3339),
		}
	}

	result := ListAllMyBucketsResult{
		Buckets: Buckets{Bucket: bucketList},
		Owner: Owner{
			ID:          strconv.FormatInt(user.ID, 10),
			DisplayName: user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(result)
}

// S3ListObjects handles S3-compatible object listing
// GET /bucket-name?prefix=...&max-keys=...
func S3ListObjects(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("S3 list objects endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bucketName := chi.URLParam(r, "bucketName")
	if bucketName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByName(bucketName, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if bucket == nil {
		// Try public buckets
		buckets, _ := bucketRepo.List(user.ID, 1000, 0)
		for _, b := range buckets {
			if b.Name == bucketName && b.IsPublic {
				bucket = b
				break
			}
		}
		if bucket == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	if bucket.UserID != user.ID && !bucket.IsPublic {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Parse query parameters
	prefix := r.URL.Query().Get("prefix")
	maxKeys := 1000
	if maxKeysStr := r.URL.Query().Get("max-keys"); maxKeysStr != "" {
		if mk, err := strconv.Atoi(maxKeysStr); err == nil && mk > 0 {
			maxKeys = mk
		}
	}

	fileRepo := db.NewFileRepository(db.GetDB())
	var files []*db.File
	if prefix != "" {
		files, err = fileRepo.ListByPrefix(bucket.ID, prefix, maxKeys, 0)
	} else {
		files, err = fileRepo.List(bucket.ID, maxKeys, 0)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to list files")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contents := make([]Content, len(files))
	for i, file := range files {
		contents[i] = Content{
			Key:          file.Name,
			LastModified: file.CreatedAt.Format(time.RFC3339),
			ETag:         fmt.Sprintf(`"%s"`, file.ETag),
			Size:         file.Size,
			StorageClass: "STANDARD",
			Owner: Owner{
				ID:          strconv.FormatInt(file.UserID, 10),
				DisplayName: "",
			},
		}
	}

	result := ListBucketResult{
		Name:        bucketName,
		Prefix:      prefix,
		MaxKeys:     maxKeys,
		IsTruncated: false,
		Contents:    contents,
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(result)
}

// S3PutObject handles S3-compatible object upload
// PUT /bucket-name/object-key
func S3PutObject(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("S3 put object endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bucketName := chi.URLParam(r, "bucketName")
	objectKey := chi.URLParam(r, "*")
	if objectKey == "" {
		objectKey = r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	}

	// URL decode the object key
	decodedKey, err := url.QueryUnescape(objectKey)
	if err == nil {
		objectKey = decodedKey
	}

	if bucketName == "" || objectKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByName(bucketName, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if bucket == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if bucket.UserID != user.ID && !bucket.IsPublic {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Create storage directory
	storageBase := getStoragePath()
	storageDir := filepath.Join(storageBase, fmt.Sprintf("%d", user.ID), fmt.Sprintf("%d", bucket.ID))
	if err := service.EnsureDir(storageDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create storage directory")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create file path
	filePath := filepath.Join(storageDir, objectKey)

	// Create file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content and calculate hash
	hash := md5.New()
	multiWriter := io.MultiWriter(dst, hash)
	size, err := io.Copy(multiWriter, r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save file")
		os.Remove(filePath)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	etag := fmt.Sprintf("%x", hash.Sum(nil))
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Save file metadata to database
	fileRepo := db.NewFileRepository(db.GetDB())
	existingFile, _ := fileRepo.GetByName(bucket.ID, objectKey)

	dbFile := &db.File{
		BucketID:    bucket.ID,
		Name:        objectKey,
		Path:        filePath,
		ContentType: contentType,
		Size:        size,
		ETag:        etag,
		UserID:      user.ID,
	}

	if existingFile != nil {
		dbFile.ID = existingFile.ID
		fileRepo.Update(dbFile)
	} else {
		fileRepo.Create(dbFile)
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
	w.WriteHeader(http.StatusOK)
}

// S3GetObject handles S3-compatible object download
// GET /bucket-name/object-key
func S3GetObject(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("S3 get object endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bucketName := chi.URLParam(r, "bucketName")
	objectKey := chi.URLParam(r, "*")
	if objectKey == "" {
		objectKey = r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	}

	// URL decode the object key
	decodedKey, err := url.QueryUnescape(objectKey)
	if err == nil {
		objectKey = decodedKey
	}

	if bucketName == "" || objectKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByName(bucketName, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if bucket == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if bucket.UserID != user.ID && !bucket.IsPublic {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	fileRepo := db.NewFileRepository(db.GetDB())
	file, err := fileRepo.GetByName(bucket.ID, objectKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if file == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !service.FileExists(file.Path) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Open file
	fileData, err := os.Open(file.Path)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fileData.Close()

	// Set headers
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, file.ETag))
	w.Header().Set("Last-Modified", file.CreatedAt.Format(http.TimeFormat))

	// Copy file to response
	_, err = io.Copy(w, fileData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send file")
		return
	}
}

// S3DeleteObject handles S3-compatible object deletion
// DELETE /bucket-name/object-key
func S3DeleteObject(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("S3 delete object endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bucketName := chi.URLParam(r, "bucketName")
	objectKey := chi.URLParam(r, "*")
	if objectKey == "" {
		objectKey = r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	}

	// URL decode the object key
	decodedKey, err := url.QueryUnescape(objectKey)
	if err == nil {
		objectKey = decodedKey
	}

	if bucketName == "" || objectKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bucketRepo := db.NewBucketRepository(db.GetDB())
	bucket, err := bucketRepo.GetByName(bucketName, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get bucket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if bucket == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if bucket.UserID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	fileRepo := db.NewFileRepository(db.GetDB())
	file, err := fileRepo.GetByName(bucket.ID, objectKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if file == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Delete file from disk
	if service.FileExists(file.Path) {
		os.Remove(file.Path)
	}

	// Delete file from database
	if err := fileRepo.Delete(file.ID); err != nil {
		log.Error().Err(err).Msg("Failed to delete file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
