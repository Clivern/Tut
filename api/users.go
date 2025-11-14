// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/clivern/tut/db"
	"github.com/clivern/tut/middleware"
	"github.com/clivern/tut/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// CreateUserRequest represents the create user request payload
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email,min=4,max=60" label:"Email"`
	Password string `json:"password" validate:"required,strong_password,min=8,max=60" label:"Password"`
	Role     string `json:"role" validate:"required,oneof=admin user readonly" label:"Role"`
	IsActive bool   `json:"isActive" label:"Is Active"`
}

// UpdateUserRequest represents the update user request payload
type UpdateUserRequest struct {
	Email    string `json:"email" validate:"required,email,min=4,max=60" label:"Email"`
	Password string `json:"password" validate:"omitempty,strong_password,min=8,max=60" label:"Password"`
	Role     string `json:"role" validate:"required,oneof=admin user readonly" label:"Role"`
	IsActive bool   `json:"isActive" label:"Is Active"`
}

// CreateUserAction handles user creation requests
func CreateUserAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Create user endpoint called")

	// Check if user is admin
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok || currentUser.Role != db.UserRoleAdmin {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Only administrators can create users",
		})
		return
	}

	var req CreateUserRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())

	// Check if user with email already exists
	existingUser, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create user",
		})
		return
	}
	if existingUser != nil {
		service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
			"errorMessage": "User with this email already exists",
		})
		return
	}

	hashedPassword, err := service.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create user",
		})
		return
	}

	user := &db.User{
		Email:       req.Email,
		Password:    hashedPassword,
		Role:        req.Role,
		APIKey:      uuid.New().String(),
		IsActive:    req.IsActive,
		LastLoginAt: time.Time{},
	}

	if err := userRepo.Create(user); err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to create user",
		})
		return
	}

	log.Info().Int64("userID", user.ID).Msg("User created successfully")
	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"isActive":    user.IsActive,
		"apiKey":      user.APIKey,
		"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// UpdateUserAction handles user update requests
func UpdateUserAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update user endpoint called")

	// Check if user is admin
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok || currentUser.Role != db.UserRoleAdmin {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Only administrators can update users",
		})
		return
	}

	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid user ID",
		})
		return
	}

	var req UpdateUserRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())

	// Get existing user
	user, err := userRepo.GetByID(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update user",
		})
		return
	}

	if user == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "User not found",
		})
		return
	}

	// Check if email is being changed and if it's already taken
	if req.Email != user.Email {
		existingUser, err := userRepo.GetByEmail(req.Email)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check existing user")
			service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": "Failed to update user",
			})
			return
		}
		if existingUser != nil && existingUser.ID != userID {
			service.WriteJSON(w, http.StatusConflict, map[string]interface{}{
				"errorMessage": "User with this email already exists",
			})
			return
		}
	}

	user.Email = req.Email
	user.Role = req.Role
	user.IsActive = req.IsActive

	// Update password only if provided
	if req.Password != "" {
		hashedPassword, err := service.HashPassword(req.Password)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash password")
			service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"errorMessage": "Failed to update user",
			})
			return
		}
		user.Password = hashedPassword
	}

	if err := userRepo.Update(user); err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update user",
		})
		return
	}

	log.Info().Int64("userID", user.ID).Msg("User updated successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":          user.ID,
		"email":       user.Email,
		"role":        user.Role,
		"isActive":    user.IsActive,
		"apiKey":      user.APIKey,
		"lastLoginAt": user.LastLoginAt.UTC().Format(time.RFC3339),
		"createdAt":   user.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   user.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

// DeleteUserAction handles user deletion requests
func DeleteUserAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Delete user endpoint called")

	// Check if user is admin
	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok || currentUser.Role != db.UserRoleAdmin {
		service.WriteJSON(w, http.StatusForbidden, map[string]interface{}{
			"errorMessage": "Only administrators can delete users",
		})
		return
	}

	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Invalid user ID",
		})
		return
	}

	// Prevent self-deletion
	if currentUser.ID == userID {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "You cannot delete your own account",
		})
		return
	}

	userRepo := db.NewUserRepository(db.GetDB())

	// Check if user exists
	user, err := userRepo.GetByID(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete user",
		})
		return
	}

	if user == nil {
		service.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
			"errorMessage": "User not found",
		})
		return
	}

	// Delete user
	if err := userRepo.Delete(userID); err != nil {
		log.Error().Err(err).Msg("Failed to delete user")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to delete user",
		})
		return
	}

	log.Info().Int64("userID", userID).Msg("User deleted successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "User deleted successfully",
	})
}
