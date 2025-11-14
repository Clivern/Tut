// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import "github.com/clivern/tut/db"

// Settings handles the application settings
type Settings struct {
	OptionRepository *db.OptionRepository
}

// SettingsOptions contains the configuration options for application settings
type SettingsOptions struct {
	ApplicationURL   string
	ApplicationEmail string
	ApplicationName  string

	MaintenanceMode bool

	SMTPServer    string
	SMTPPort      string
	SMTPFromEmail string
	SMTPUsername  string
	SMTPPassword  string
	SMTPUseTLS    bool
}

// NewSettings creates a new Settings instance with the provided repository
func NewSettings(optionRepository *db.OptionRepository) *Settings {
	return &Settings{OptionRepository: optionRepository}
}

// UpdateSettings updates the application settings
func (s *Settings) UpdateSettings(options *SettingsOptions) error {
	err := s.OptionRepository.Update("app_url", options.ApplicationURL)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Update("app_email", options.ApplicationEmail)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Update("app_name", options.ApplicationName)
	if err != nil {
		return err
	}

	maintenanceModeStr := "0"
	if options.MaintenanceMode {
		maintenanceModeStr = "1"
	}
	err = s.OptionRepository.Update("maintenance_mode", maintenanceModeStr)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Update("smtp_server", options.SMTPServer)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Update("smtp_port", options.SMTPPort)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Update("smtp_from_email", options.SMTPFromEmail)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Update("smtp_username", options.SMTPUsername)
	if err != nil {
		return err
	}

	err = s.OptionRepository.Update("smtp_password", options.SMTPPassword)
	if err != nil {
		return err
	}

	smtpUseTLSStr := "0"
	if options.SMTPUseTLS {
		smtpUseTLSStr = "1"
	}
	err = s.OptionRepository.Update("smtp_use_tls", smtpUseTLSStr)
	if err != nil {
		return err
	}

	return nil
}

// GetSettings retrieves the application settings
func (s *Settings) GetSettings() (*SettingsOptions, error) {
	settings := &SettingsOptions{
		ApplicationURL:   "",
		ApplicationEmail: "",
		ApplicationName:  "",
		MaintenanceMode:  false,
		SMTPServer:       "",
		SMTPPort:         "",
		SMTPFromEmail:    "",
		SMTPUsername:     "",
		SMTPPassword:     "",
		SMTPUseTLS:       false,
	}
	option, err := s.OptionRepository.Get("app_url")
	if err != nil {
		return nil, err
	}
	settings.ApplicationURL = option.Value

	option, err = s.OptionRepository.Get("app_email")
	if err != nil {
		return nil, err
	}
	settings.ApplicationEmail = option.Value

	option, err = s.OptionRepository.Get("app_name")
	if err != nil {
		return nil, err
	}
	settings.ApplicationName = option.Value

	option, err = s.OptionRepository.Get("maintenance_mode")
	if err != nil {
		return nil, err
	}
	settings.MaintenanceMode = option.Value == "1"

	option, err = s.OptionRepository.Get("smtp_server")
	if err != nil {
		return nil, err
	}
	settings.SMTPServer = option.Value

	option, err = s.OptionRepository.Get("smtp_port")
	if err != nil {
		return nil, err
	}
	settings.SMTPPort = option.Value

	option, err = s.OptionRepository.Get("smtp_from_email")
	if err != nil {
		return nil, err
	}
	settings.SMTPFromEmail = option.Value

	option, err = s.OptionRepository.Get("smtp_username")
	if err != nil {
		return nil, err
	}
	settings.SMTPUsername = option.Value

	option, err = s.OptionRepository.Get("smtp_password")
	if err != nil {
		return nil, err
	}
	settings.SMTPPassword = option.Value

	option, err = s.OptionRepository.Get("smtp_use_tls")
	if err != nil {
		return nil, err
	}
	settings.SMTPUseTLS = option.Value == "1"

	return settings, nil
}
