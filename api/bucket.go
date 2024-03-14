// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/clivern/tut/db"
	"github.com/clivern/tut/middleware"
	"github.com/clivern/tut/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateBucketRequest represents the request payload for creating a bucket
type CreateBucketRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=63,alphanum"`
	Description string `json:"description" validate:"max=500"`
	IsPublic    bool   `json:"is_public"`
}

// CreateBucket creates a new bucket
func CreateBucket(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Create bucket endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	var req CreateBucketRequest
	if err := service.DecodeJSON(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	if err := service.ValidateStruct(req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	bucketRepo := db.NewBucketRepository(db.GetDB())

	// Check if bucket with same name already exists for this user
	existing, err := bucketRepo.GetByName(req.Name, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create bucket",
		})
		return
	}
	if existing != nil {
		service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": "Bucket with this name already exists",
		})
		return
	}

	bucket := &db.Bucket{
		Name:        req.Name,
		UserID:      user.ID,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	if err := bucketRepo.Create(bucket); err != nil {
		log.Error().Err(err).Msg("Failed to create bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create bucket",
		})
		return
	}

	log.Info().
		Int64("bucket_id", bucket.ID).
		Str("bucket_name", bucket.Name).
		Int64("user_id", user.ID).
		Msg("Bucket created successfully")

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          bucket.ID,
		"name":        bucket.Name,
		"description": bucket.Description,
		"is_public":   bucket.IsPublic,
		"created_at":  bucket.CreatedAt,
		"updated_at":  bucket.UpdatedAt,
	})
}

// ListBuckets lists all buckets for the authenticated user
func ListBuckets(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List buckets endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
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

	bucketRepo := db.NewBucketRepository(db.GetDB())
	buckets, err := bucketRepo.List(user.ID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list buckets")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list buckets",
		})
		return
	}

	total, err := bucketRepo.Count(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count buckets")
	}

	result := make([]map[string]interface{}, len(buckets))
	for i, bucket := range buckets {
		result[i] = map[string]interface{}{
			"id":          bucket.ID,
			"name":        bucket.Name,
			"description": bucket.Description,
			"is_public":   bucket.IsPublic,
			"created_at":  bucket.CreatedAt,
			"updated_at":  bucket.UpdatedAt,
		}
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"buckets": result,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetBucket retrieves a bucket by ID
func GetBucket(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get bucket endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketIDStr := chi.URLParam(r, "id")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

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

	// Check if user owns the bucket or bucket is public
	if bucket.UserID != user.ID && !bucket.IsPublic {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":          bucket.ID,
		"name":        bucket.Name,
		"description": bucket.Description,
		"is_public":   bucket.IsPublic,
		"created_at":  bucket.CreatedAt,
		"updated_at":  bucket.UpdatedAt,
	})
}

// DeleteBucket deletes a bucket
func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Delete bucket endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Unauthorized",
		})
		return
	}

	bucketIDStr := chi.URLParam(r, "id")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

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

	// Only bucket owner can delete
	if bucket.UserID != user.ID {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Access denied",
		})
		return
	}

	// Check if bucket has files
	fileRepo := db.NewFileRepository(db.GetDB())
	fileCount, err := fileRepo.Count(bucketID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count files")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete bucket",
		})
		return
	}

	if fileCount > 0 {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Cannot delete bucket with existing files",
		})
		return
	}

	if err := bucketRepo.Delete(bucketID); err != nil {
		log.Error().Err(err).Msg("Failed to delete bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete bucket",
		})
		return
	}

	log.Info().
		Int64("bucket_id", bucketID).
		Int64("user_id", user.ID).
		Msg("Bucket deleted successfully")

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Bucket deleted successfully",
	})
}
