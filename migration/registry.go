// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package migration

import (
	"database/sql"
	"fmt"
	"strings"
)

// detectDriver attempts to determine the database driver type
func detectDriver(db *sql.DB) string {
	// Check SQLite
	_, err := db.Exec("SELECT sqlite_version()")
	if err == nil {
		return "sqlite"
	}

	// Check PostgreSQL
	_, err = db.Exec("SELECT version()")
	if err == nil {
		var version string
		db.QueryRow("SELECT version()").Scan(&version)
		if strings.Contains(strings.ToLower(version), "postgresql") {
			return "postgres"
		}
	}

	// Unknown database driver
	return "unknown"
}

// GetAll returns all registered migrations
func GetAll() []Migration {
	return []Migration{
		{
			Version:     "20250101000003",
			Description: "Create options table",
			Up:          createOptionsTable,
			Down:        dropOptionsTable,
		},
		{
			Version:     "20250101000004",
			Description: "Create users table",
			Up:          createUsersTable,
			Down:        dropUsersTable,
		},
		{
			Version:     "20250101000005",
			Description: "Create users_meta table",
			Up:          createUsersMetaTable,
			Down:        dropUsersMetaTable,
		},
		{
			Version:     "20250101000006",
			Description: "Create sessions table",
			Up:          createSessionsTable,
			Down:        dropSessionsTable,
		},
		{
			Version:     "20250101000007",
			Description: "Create activities table",
			Up:          createActivitiesTable,
			Down:        dropActivitiesTable,
		},
	}
}

// createOptionsTable creates the options table
func createOptionsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE options (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key VARCHAR(255) NOT NULL UNIQUE,
			value TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE options (
			id SERIAL PRIMARY KEY,
			key VARCHAR(255) NOT NULL UNIQUE,
			value TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_key ON options(key)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropOptionsTable drops the options table
func dropOptionsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS options")
	return err
}

// createUsersTable creates the users table
func createUsersTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		// role is admin, user or readonly
		query = `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			api_key VARCHAR(255) UNIQUE,
			is_active BOOLEAN DEFAULT 1,
			last_login_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgres":
		query = `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			api_key VARCHAR(255) UNIQUE,
			is_active BOOLEAN DEFAULT true,
			last_login_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_email ON users(email);
		CREATE INDEX idx_api_key ON users(api_key)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUsersTable drops the users table
func dropUsersTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	return err
}

// createUsersMetaTable creates the users_meta table
func createUsersMetaTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE users_meta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key VARCHAR(255) NOT NULL,
			value TEXT,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(user_id, key)
		)`
	case "postgres":
		query = `
		CREATE TABLE users_meta (
			id SERIAL PRIMARY KEY,
			key VARCHAR(255) NOT NULL,
			value TEXT,
			user_id INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE (user_id, key)
		);
		CREATE INDEX idx_user_id ON users_meta(user_id);
		CREATE INDEX idx_key ON users_meta(key)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropUsersMetaTable drops the users_meta table
func dropUsersMetaTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS users_meta")
	return err
}

// createSessionsTable creates the sessions table
func createSessionsTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token VARCHAR(255) NOT NULL UNIQUE,
			user_id INTEGER NOT NULL,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`
	case "postgres":
		query = `
		CREATE TABLE sessions (
			id BIGSERIAL PRIMARY KEY,
			token VARCHAR(255) NOT NULL UNIQUE,
			user_id INT NOT NULL,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX idx_token ON sessions(token);
		CREATE INDEX idx_user_id ON sessions(user_id);
		CREATE INDEX idx_expires_at ON sessions(expires_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropSessionsTable drops the sessions table
func dropSessionsTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS sessions")
	return err
}

// createActivitiesTable creates the activities table
func createActivitiesTable(db *sql.DB) error {
	driver := detectDriver(db)
	var query string

	switch driver {
	case "sqlite":
		query = `
		CREATE TABLE activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			user_email VARCHAR(255),
			action VARCHAR(100) NOT NULL,
			entity_type VARCHAR(50) NOT NULL,
			entity_id INTEGER,
			details TEXT,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`
	case "postgres":
		query = `
		CREATE TABLE activities (
			id BIGSERIAL PRIMARY KEY,
			user_id INT,
			user_email VARCHAR(255),
			action VARCHAR(100) NOT NULL,
			entity_type VARCHAR(50) NOT NULL,
			entity_id INT,
			details TEXT,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		);
		CREATE INDEX idx_user_id ON activities(user_id);
		CREATE INDEX idx_action ON activities(action);
		CREATE INDEX idx_entity ON activities(entity_type, entity_id);
		CREATE INDEX idx_created_at ON activities(created_at)`
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	_, err := db.Exec(query)
	return err
}

// dropActivitiesTable drops the activities table
func dropActivitiesTable(db *sql.DB) error {
	_, err := db.Exec("DROP TABLE IF EXISTS activities")
	return err
}
