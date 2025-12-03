// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/clivern/tut/db"
	"github.com/clivern/tut/module"
	"github.com/clivern/tut/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// CreateBucketRequest represents the create bucket request payload
type CreateBucketRequest struct {
	Name         string            `json:"name" validate:"required,bucket_name" label:"Bucket Name"`
	Region       string            `json:"region" validate:"omitempty,region" label:"Region"`
	Tags         map[string]string `json:"tags" label:"Tags"`
	Versioning   bool              `json:"versioning" label:"Versioning"`
	PublicAccess bool              `json:"publicAccess" label:"Public Access"`
}

// UpdateBucketRequest represents the update bucket request payload
type UpdateBucketRequest struct {
	Name         string            `json:"name" validate:"omitempty,bucket_name" label:"Bucket Name"`
	Region       string            `json:"region" validate:"omitempty,region" label:"Region"`
	Tags         map[string]string `json:"tags" label:"Tags"`
	Versioning   *bool             `json:"versioning" label:"Versioning"`
	PublicAccess *bool             `json:"publicAccess" label:"Public Access"`
}

// CreateBucketAction handles bucket creation requests
func CreateBucketAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Create bucket endpoint called")

	var req CreateBucketRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	bucketModule := module.NewBucket(db.NewBucketRepository(db.GetDB()))
	bucket, err := bucketModule.CreateBucket(&module.CreateBucketOptions{
		Name:         req.Name,
		Region:       req.Region,
		Tags:         req.Tags,
		Versioning:   req.Versioning,
		PublicAccess: req.PublicAccess,
	})

	if err != nil {
		if errors.Is(err, module.ErrBucketNameAlreadyExists) {
			service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": "Bucket with this name already exists",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to create bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create bucket",
		})
		return
	}

	log.Info().Int64("bucketID", bucket.ID).Str("bucketName", bucket.Name).Msg("Bucket created successfully")
	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           bucket.ID,
		"name":         bucket.Name,
		"region":       bucket.Region,
		"tags":         bucket.Tags,
		"versioning":   bucket.Versioning,
		"publicAccess": bucket.PublicAccess,
		"createdAt":    bucket.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":    bucket.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetBucketAction handles get bucket by ID requests
func GetBucketAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get bucket endpoint called")

	bucketIDStr := chi.URLParam(r, "id")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	bucketModule := module.NewBucket(db.NewBucketRepository(db.GetDB()))
	bucket, err := bucketModule.GetBucketByID(bucketID)
	if err != nil {
		if errors.Is(err, module.ErrBucketNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get bucket",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":           bucket.ID,
		"name":         bucket.Name,
		"region":       bucket.Region,
		"tags":         bucket.Tags,
		"versioning":   bucket.Versioning,
		"publicAccess": bucket.PublicAccess,
		"createdAt":    bucket.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":    bucket.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateBucketAction handles bucket update requests
func UpdateBucketAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update bucket endpoint called")

	bucketIDStr := chi.URLParam(r, "id")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	var req UpdateBucketRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	bucketModule := module.NewBucket(db.NewBucketRepository(db.GetDB()))
	bucket, err := bucketModule.UpdateBucketByID(bucketID, &module.UpdateBucketOptions{
		Name:         req.Name,
		Region:       req.Region,
		Tags:         req.Tags,
		Versioning:   req.Versioning,
		PublicAccess: req.PublicAccess,
	})

	if err != nil {
		if errors.Is(err, module.ErrBucketNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket not found",
			})
			return
		}
		if errors.Is(err, module.ErrBucketNameAlreadyExists) {
			service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": "Bucket with this name already exists",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update bucket",
		})
		return
	}

	log.Info().Int64("bucketID", bucket.ID).Str("bucketName", bucket.Name).Msg("Bucket updated successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":           bucket.ID,
		"name":         bucket.Name,
		"region":       bucket.Region,
		"tags":         bucket.Tags,
		"versioning":   bucket.Versioning,
		"publicAccess": bucket.PublicAccess,
		"createdAt":    bucket.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":    bucket.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListBucketsAction handles bucket listing requests with pagination
func ListBucketsAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List buckets endpoint called")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	bucketModule := module.NewBucket(db.NewBucketRepository(db.GetDB()))
	result, err := bucketModule.ListBuckets(&module.ListBucketsOptions{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to list buckets")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list buckets",
		})
		return
	}

	bucketList := make([]map[string]interface{}, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		bucketList = append(bucketList, map[string]interface{}{
			"id":           bucket.ID,
			"name":         bucket.Name,
			"region":       bucket.Region,
			"tags":         bucket.Tags,
			"versioning":   bucket.Versioning,
			"publicAccess": bucket.PublicAccess,
			"createdAt":    bucket.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			"updatedAt":    bucket.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"buckets": bucketList,
		"_meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// DeleteBucketAction handles bucket deletion requests
func DeleteBucketAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Delete bucket endpoint called")

	bucketIDStr := chi.URLParam(r, "id")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	bucketModule := module.NewBucket(db.NewBucketRepository(db.GetDB()))
	err = bucketModule.DeleteBucketByID(bucketID)
	if err != nil {
		if errors.Is(err, module.ErrBucketNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to delete bucket")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete bucket",
		})
		return
	}

	log.Info().Int64("bucketID", bucketID).Msg("Bucket deleted successfully")
	service.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
