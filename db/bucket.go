// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Bucket represents a storage bucket in the database.
type Bucket struct {
	ID          int64
	Name        string
	UserID      int64
	Description string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BucketRepository handles database operations for buckets.
type BucketRepository struct {
	db *sql.DB
}

// NewBucketRepository creates a new bucket repository.
func NewBucketRepository(db *sql.DB) *BucketRepository {
	return &BucketRepository{db: db}
}

// Create inserts a new bucket into the database.
func (r *BucketRepository) Create(bucket *Bucket) error {
	result, err := r.db.Exec(
		`INSERT INTO buckets (name, user_id, description, is_public)
		VALUES (?, ?, ?, ?)`,
		bucket.Name,
		bucket.UserID,
		bucket.Description,
		bucket.IsPublic,
	)
	if err != nil {
		return err
	}

	bucket.ID, err = result.LastInsertId()
	return err
}

// GetByID retrieves a bucket by ID.
func (r *BucketRepository) GetByID(id int64) (*Bucket, error) {
	bucket := &Bucket{}
	err := r.db.QueryRow(
		`SELECT id, name, user_id, description, is_public, created_at, updated_at
		FROM buckets
		WHERE id = ?`,
		id,
	).Scan(
		&bucket.ID,
		&bucket.Name,
		&bucket.UserID,
		&bucket.Description,
		&bucket.IsPublic,
		&bucket.CreatedAt,
		&bucket.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

// GetByName retrieves a bucket by name for a specific user.
func (r *BucketRepository) GetByName(name string, userID int64) (*Bucket, error) {
	bucket := &Bucket{}
	err := r.db.QueryRow(
		`SELECT id, name, user_id, description, is_public, created_at, updated_at
		FROM buckets
		WHERE name = ? AND user_id = ?`,
		name,
		userID,
	).Scan(
		&bucket.ID,
		&bucket.Name,
		&bucket.UserID,
		&bucket.Description,
		&bucket.IsPublic,
		&bucket.CreatedAt,
		&bucket.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

// Update updates a bucket's information.
func (r *BucketRepository) Update(bucket *Bucket) error {
	_, err := r.db.Exec(
		`UPDATE buckets SET
			name = ?, description = ?, is_public = ?, updated_at = ?
		WHERE id = ?`,
		bucket.Name,
		bucket.Description,
		bucket.IsPublic,
		time.Now().UTC(),
		bucket.ID,
	)
	return err
}

// Delete removes a bucket from the database.
func (r *BucketRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM buckets WHERE id = ?", id)
	return err
}

// List retrieves all buckets for a user with pagination.
func (r *BucketRepository) List(userID int64, limit, offset int) ([]*Bucket, error) {
	rows, err := r.db.Query(
		`SELECT id, name, user_id, description, is_public, created_at, updated_at
		FROM buckets
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []*Bucket
	for rows.Next() {
		bucket := &Bucket{}
		if err := rows.Scan(
			&bucket.ID,
			&bucket.Name,
			&bucket.UserID,
			&bucket.Description,
			&bucket.IsPublic,
			&bucket.CreatedAt,
			&bucket.UpdatedAt,
		); err != nil {
			return nil, err
		}
		buckets = append(buckets, bucket)
	}

	return buckets, rows.Err()
}

// Count returns the total number of buckets for a user.
func (r *BucketRepository) Count(userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM buckets WHERE user_id = ?", userID).Scan(&count)
	return count, err
}
