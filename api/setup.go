// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/clivern/tut/db"
	"github.com/clivern/tut/module"
	"github.com/clivern/tut/service"

	"github.com/rs/zerolog/log"
)

// SetupRequest represents the setup request payload
type SetupRequest struct {
	ApplicationURL   string `json:"applicationURL" validate:"required,url,min=4,max=60" label:"Application URL"`
	ApplicationEmail string `json:"applicationEmail" validate:"required,email,min=4,max=60" label:"Application Email"`
	ApplicationName  string `json:"applicationName" validate:"required,min=2,max=50" label:"Application Name"`
	AdminEmail       string `json:"adminEmail" validate:"required,email,min=4,max=60" label:"Admin Email"`
	AdminPassword    string `json:"adminPassword" validate:"required,strong_password,min=8,max=60" label:"Admin Password"`
}

// SetupAction handles the setup installation
func SetupAction(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest

	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	setupModule := module.NewSetup(
		db.NewOptionRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)

	if setupModule.IsInstalled() {
		service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "Application is already installed",
		})
		return
	}

	err := setupModule.Install(&module.SetupOptions{
		ApplicationURL:   req.ApplicationURL,
		ApplicationEmail: req.ApplicationEmail,
		ApplicationName:  req.ApplicationName,
		AdminEmail:       req.AdminEmail,
		AdminPassword:    req.AdminPassword,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to complete setup")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to complete setup",
		})
		return
	}

	log.Info().Msg("Application setup completed successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "Application setup completed successfully",
	})
}

// SetupStatusAction checks if the application is already installed
func SetupStatusAction(w http.ResponseWriter, _ *http.Request) {
	setupModule := module.NewSetup(
		db.NewOptionRepository(db.GetDB()),
		db.NewUserRepository(db.GetDB()),
	)
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"installed": setupModule.IsInstalled(),
	})
}
