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

// SettingsRequest represents the settings request payload
type SettingsRequest struct {
	ApplicationURL   string `json:"applicationURL" validate:"required,url,min=4,max=60" label:"Application URL"`
	ApplicationEmail string `json:"applicationEmail" validate:"required,email,min=4,max=60" label:"Application Email"`
	ApplicationName  string `json:"applicationName" validate:"required,min=2,max=50" label:"Application Name"`
	MaintenanceMode  bool   `json:"maintenanceMode" validate:"required,boolean" label:"Maintenance Mode"`
	SMTPServer       string `json:"smtpServer" validate:"required,min=4,max=60" label:"SMTP Server"`
	SMTPPort         string `json:"smtpPort" validate:"required,min=1,max=5" label:"SMTP Port"`
	SMTPFromEmail    string `json:"smtpFromEmail" validate:"required,email,min=4,max=60" label:"SMTP From Email"`
	SMTPUsername     string `json:"smtpUsername" validate:"required,min=4,max=60" label:"SMTP Username"`
	SMTPPassword     string `json:"smtpPassword" validate:"required,min=8,max=60" label:"SMTP Password"`
	SMTPUseTLS       bool   `json:"smtpUseTLS" validate:"required,boolean" label:"SMTP Use TLS"`
}

// UpdateSettingsAction handles user settings update requests
func UpdateSettingsAction(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Update settings endpoint called")

	var req SettingsRequest
	if err := service.DecodeAndValidate(r, &req); err != nil {
		service.WriteValidationError(w, err)
		return
	}

	settingsModule := module.NewSettings(db.NewOptionRepository(db.GetDB()))
	err := settingsModule.UpdateSettings(&module.SettingsOptions{
		ApplicationURL:   req.ApplicationURL,
		ApplicationEmail: req.ApplicationEmail,
		ApplicationName:  req.ApplicationName,
		MaintenanceMode:  req.MaintenanceMode,
		SMTPServer:       req.SMTPServer,
		SMTPPort:         req.SMTPPort,
		SMTPFromEmail:    req.SMTPFromEmail,
		SMTPUsername:     req.SMTPUsername,
		SMTPPassword:     req.SMTPPassword,
		SMTPUseTLS:       req.SMTPUseTLS,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to update settings")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to update settings",
		})
		return
	}

	log.Info().Msg("Settings updated successfully")
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"successMessage": "Settings updated successfully",
	})
}

// GetSettingsAction handles user settings get requests
func GetSettingsAction(w http.ResponseWriter, _ *http.Request) {
	log.Debug().Msg("Get settings endpoint called")

	settingsModule := module.NewSettings(db.NewOptionRepository(db.GetDB()))
	settings, err := settingsModule.GetSettings()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get settings")
		service.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": "Failed to get settings",
		})
		return
	}
	service.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"settings": settings,
	})
}
