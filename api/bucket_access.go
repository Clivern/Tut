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

// CreateBucketAccessRequest represents the create bucket access request payload
type CreateBucketAccessRequest struct {
	UserID     int64  `json:"userId" validate:"required" label:"User ID"`
	Permission string `json:"permission" validate:"required,oneof=read write admin" label:"Permission"`
}

// UpdateBucketAccessRequest represents the update bucket access request payload
type UpdateBucketAccessRequest struct {
	Permission string `json:"permission" validate:"required,oneof=read write admin" label:"Permission"`
}

// CreateBucketAccessAction handles bucket access creation requests
func CreateBucketAccessAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Create bucket access endpoint called")

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	var req CreateBucketAccessRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	bucketAccessModule := module.NewBucketAccess(
		db.NewBucketAccessRepository(db.GetDB()),
		db.NewBucketRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)

	access, err := bucketAccessModule.CreateBucketAccess(&module.CreateBucketAccessOptions{
		BucketID:   bucketID,
		UserID:     req.UserID,
		Permission: req.Permission,
	})

	if err != nil {
		if errors.Is(err, module.ErrBucketNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket not found",
			})
			return
		}
		if errors.Is(err, module.ErrUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User not found",
			})
			return
		}
		if errors.Is(err, module.ErrInvalidPermission) {
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": "Invalid permission. Must be one of: read, write, admin",
			})
			return
		}
		if errors.Is(err, module.ErrBucketAccessAlreadyExists) {
			service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": "Access for this user already exists for this bucket",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to create bucket access")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create bucket access",
		})
		return
	}

	log.Info().
		Int64("bucketID", access.BucketID).
		Int64("userID", access.UserID).
		Str("permission", access.Permission).
		Msg("Bucket access created successfully")

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         access.ID,
		"bucketId":   access.BucketID,
		"userId":     access.UserID,
		"permission": access.Permission,
		"createdAt":  access.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":  access.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListBucketAccessAction handles bucket access listing requests
func ListBucketAccessAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("List bucket access endpoint called")

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	bucketAccessModule := module.NewBucketAccess(
		db.NewBucketAccessRepository(db.GetDB()),
		db.NewBucketRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)

	result, err := bucketAccessModule.ListBucketAccess(&module.ListBucketAccessOptions{
		BucketID: bucketID,
	})

	if err != nil {
		if errors.Is(err, module.ErrBucketNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to list bucket access")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to list bucket access",
		})
		return
	}

	accessList := make([]map[string]interface{}, 0, len(result.Accesses))
	for _, access := range result.Accesses {
		accessList = append(accessList, map[string]interface{}{
			"id":         access.ID,
			"bucketId":   access.BucketID,
			"userId":     access.UserID,
			"permission": access.Permission,
			"createdAt":  access.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			"updatedAt":  access.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"access": accessList,
	})
}

// GetBucketAccessAction handles get bucket access by user ID requests
func GetBucketAccessAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get bucket access endpoint called")

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid user ID",
		})
		return
	}

	bucketAccessModule := module.NewBucketAccess(
		db.NewBucketAccessRepository(db.GetDB()),
		db.NewBucketRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)

	access, err := bucketAccessModule.GetBucketAccess(bucketID, userID)
	if err != nil {
		if errors.Is(err, module.ErrBucketAccessNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket access not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get bucket access")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get bucket access",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":         access.ID,
		"bucketId":   access.BucketID,
		"userId":     access.UserID,
		"permission": access.Permission,
		"createdAt":  access.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":  access.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateBucketAccessAction handles bucket access update requests
func UpdateBucketAccessAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update bucket access endpoint called")

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid user ID",
		})
		return
	}

	var req UpdateBucketAccessRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	bucketAccessModule := module.NewBucketAccess(
		db.NewBucketAccessRepository(db.GetDB()),
		db.NewBucketRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)

	access, err := bucketAccessModule.UpdateBucketAccess(bucketID, userID, &module.UpdateBucketAccessOptions{
		Permission: req.Permission,
	})

	if err != nil {
		if errors.Is(err, module.ErrBucketAccessNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket access not found",
			})
			return
		}
		if errors.Is(err, module.ErrInvalidPermission) {
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"errorMessage": "Invalid permission. Must be one of: read, write, admin",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update bucket access")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update bucket access",
		})
		return
	}

	log.Info().
		Int64("bucketID", access.BucketID).
		Int64("userID", access.UserID).
		Str("permission", access.Permission).
		Msg("Bucket access updated successfully")

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":         access.ID,
		"bucketId":   access.BucketID,
		"userId":     access.UserID,
		"permission": access.Permission,
		"createdAt":  access.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt":  access.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteBucketAccessAction handles bucket access deletion requests
func DeleteBucketAccessAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Delete bucket access endpoint called")

	bucketIDStr := chi.URLParam(r, "bucketId")
	bucketID, err := strconv.ParseInt(bucketIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid bucket ID",
		})
		return
	}

	userIDStr := chi.URLParam(r, "userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid user ID",
		})
		return
	}

	bucketAccessModule := module.NewBucketAccess(
		db.NewBucketAccessRepository(db.GetDB()),
		db.NewBucketRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)

	err = bucketAccessModule.DeleteBucketAccess(bucketID, userID)
	if err != nil {
		if errors.Is(err, module.ErrBucketAccessNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "Bucket access not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to delete bucket access")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete bucket access",
		})
		return
	}

	log.Info().
		Int64("bucketID", bucketID).
		Int64("userID", userID).
		Msg("Bucket access deleted successfully")

	service.WriteJSON(w, http.StatusNoContent, map[string]interface{}{})
}
