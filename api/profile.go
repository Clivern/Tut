// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/clivern/tut/db"
	"github.com/clivern/tut/middleware"
	"github.com/clivern/tut/module"
	"github.com/clivern/tut/service"

	"github.com/rs/zerolog/log"
)

// GetProfileAction handles user profile requests
func GetProfileAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get profile endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	gravatar := &service.Gravatar{}
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"role":        user.Role,
			"avatar":      gravatar.GetGravatar(user.Email, 200),
			"isActive":    user.IsActive,
			"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
		},
	})
}

// UpdateProfileRequest represents the update profile request payload
type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=1,max=100" label:"Name"`
	Email string `json:"email" validate:"required,email,min=4,max=60" label:"Email"`
}

// UpdateProfileAction handles user profile update requests
func UpdateProfileAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update profile endpoint called")

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	var req UpdateProfileRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	updatedUser, err := userModule.UpdateUser(&module.UpdateUserOptions{
		UserID:   user.ID,
		Name:     req.Name,
		Email:    req.Email,
		Password: "",            // Don't update password
		Role:     user.Role,     // Keep existing role
		IsActive: user.IsActive, // Keep existing status
	})

	if err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User not found",
			})
			return
		}
		if errors.Is(err, module.ErrUserEmailAlreadyExists) {
			service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": "User with this email already exists",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update profile")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update profile",
		})
		return
	}

	gravatar := &service.Gravatar{}
	log.Info().Int64("userID", user.ID).Msg("Profile updated successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":          updatedUser.ID,
			"name":        updatedUser.Name,
			"email":       updatedUser.Email,
			"role":        updatedUser.Role,
			"avatar":      gravatar.GetGravatar(updatedUser.Email, 200),
			"isActive":    updatedUser.IsActive,
			"lastLoginAt": updatedUser.LastLoginAt.UTC().Format(time.RFC3339),
			"createdAt":   updatedUser.CreatedAt.UTC().Format(time.RFC3339),
			"updatedAt":   updatedUser.UpdatedAt.UTC().Format(time.RFC3339),
		},
		"successMessage": "Profile updated successfully",
	})
}

// GetAPIKeyAction handles API key fetch requests (returns full key)
func GetAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Get API key endpoint called")

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	user, err := userModule.GetUser(currentUser.ID)
	if err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get API key",
		})
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"apiKey": user.APIKey,
	})
}

// RotateAPIKeyAction handles API key rotation requests (regenerates and returns new key)
func RotateAPIKeyAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Rotate API key endpoint called")

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	newAPIKey, err := userModule.RegenerateAPIKey(currentUser.ID)
	if err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to rotate API key")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to rotate API key",
		})
		return
	}

	log.Info().Int64("userID", currentUser.ID).Msg("API key rotated successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"apiKey":  newAPIKey,
		"message": "API key rotated successfully",
	})
}

// UpdatePasswordRequest represents the update password request payload
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required" label:"Current Password"`
	NewPassword     string `json:"newPassword" validate:"required,strong_password,min=8,max=60" label:"New Password"`
}

// UpdatePasswordAction handles password update requests
func UpdatePasswordAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update password endpoint called")

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		service.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"errorMessage": "Not authenticated",
		})
		return
	}

	var req UpdatePasswordRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	// Fetch fresh user data to verify current password
	userModule := module.NewUser(db.NewUserRepository(db.GetDB()))
	user, err := userModule.GetUser(currentUser.ID)
	if err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update password",
		})
		return
	}

	// Verify current password
	if !service.ComparePassword(user.Password, req.CurrentPassword) {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Current password is incorrect",
		})
		return
	}

	// Update password
	if err := userModule.UpdatePassword(currentUser.ID, req.NewPassword); err != nil {
		if errors.Is(err, module.ErrUserNotFound) {
			service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
				"errorMessage": "User not found",
			})
			return
		}
		log.Error().Err(err).Msg("Failed to update password")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update password",
		})
		return
	}

	log.Info().Int64("userID", currentUser.ID).Msg("Password updated successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Password updated successfully",
	})
}
