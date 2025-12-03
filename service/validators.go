// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package service

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validateStrongPassword validates that a password meets security requirements
// Requires: min 8 chars, 1 uppercase, 1 lowercase, 1 digit, 1 special character
//
// Usage: Password string `validate:"strong_password"`
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}
	// Has uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Has lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Has digit
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Has special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// validateBucketName validates a bucket name according to common S3-compatible rules.
// Bucket names must be 3-63 characters, lowercase, and can only contain letters, numbers, dots, and hyphens.
//
// Usage: Name string `validate:"bucket_name"`
func validateBucketName(fl validator.FieldLevel) bool {
	name := strings.TrimSpace(fl.Field().String())
	if name == "" {
		return false
	}

	// Bucket names must be 3-63 characters long
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	// Bucket names must be lowercase
	if name != strings.ToLower(name) {
		return false
	}

	// Bucket names can only contain lowercase letters, numbers, dots, and hyphens
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.' ||
			char == '-') {
			return false
		}
	}

	// Bucket names cannot start or end with a dot or hyphen
	if name[0] == '.' || name[0] == '-' ||
		name[len(name)-1] == '.' || name[len(name)-1] == '-' {
		return false
	}

	// Bucket names cannot contain consecutive dots
	if strings.Contains(name, "..") {
		return false
	}

	return true
}
