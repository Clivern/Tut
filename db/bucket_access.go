// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Bucket access permission constants
const (
	BucketPermissionRead  = "read"
	BucketPermissionWrite = "write"
	BucketPermissionAdmin = "admin"
)

// BucketAccess represents a user's access permission to a bucket.
type BucketAccess struct {
	ID         int64
	BucketID   int64
	UserID     int64
	Permission string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// BucketAccessRepository handles database operations for bucket access.
type BucketAccessRepository struct {
	db *sql.DB
}

// NewBucketAccessRepository creates a new bucket access repository.
func NewBucketAccessRepository(db *sql.DB) *BucketAccessRepository {
	return &BucketAccessRepository{db: db}
}

// Create inserts a new bucket access record into the database.
func (r *BucketAccessRepository) Create(access *BucketAccess) error {
	result, err := r.db.Exec(
		`INSERT INTO bucket_access (bucket_id, user_id, permission)
		VALUES (?, ?, ?)`,
		access.BucketID,
		access.UserID,
		access.Permission,
	)
	if err != nil {
		return err
	}

	access.ID, err = result.LastInsertId()
	return err
}

// Get retrieves a bucket access record by bucket ID and user ID.
func (r *BucketAccessRepository) Get(bucketID, userID int64) (*BucketAccess, error) {
	access := &BucketAccess{}
	err := r.db.QueryRow(
		`SELECT id, bucket_id, user_id, permission, created_at, updated_at
		FROM bucket_access
		WHERE bucket_id = ? AND user_id = ?`,
		bucketID,
		userID,
	).Scan(
		&access.ID,
		&access.BucketID,
		&access.UserID,
		&access.Permission,
		&access.CreatedAt,
		&access.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return access, nil
}

// GetByID retrieves a bucket access record by ID.
func (r *BucketAccessRepository) GetByID(id int64) (*BucketAccess, error) {
	access := &BucketAccess{}
	err := r.db.QueryRow(
		`SELECT id, bucket_id, user_id, permission, created_at, updated_at
		FROM bucket_access
		WHERE id = ?`,
		id,
	).Scan(
		&access.ID,
		&access.BucketID,
		&access.UserID,
		&access.Permission,
		&access.CreatedAt,
		&access.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return access, nil
}

// Update updates a bucket access record's permission.
func (r *BucketAccessRepository) Update(access *BucketAccess) error {
	_, err := r.db.Exec(
		`UPDATE bucket_access SET
			permission = ?, updated_at = ?
		WHERE id = ?`,
		access.Permission,
		time.Now().UTC(),
		access.ID,
	)
	return err
}

// Delete removes a bucket access record from the database.
func (r *BucketAccessRepository) Delete(bucketID, userID int64) error {
	_, err := r.db.Exec(
		"DELETE FROM bucket_access WHERE bucket_id = ? AND user_id = ?",
		bucketID,
		userID,
	)
	return err
}

// DeleteByID removes a bucket access record by ID.
func (r *BucketAccessRepository) DeleteByID(id int64) error {
	_, err := r.db.Exec("DELETE FROM bucket_access WHERE id = ?", id)
	return err
}

// ListByBucket retrieves all access records for a bucket.
func (r *BucketAccessRepository) ListByBucket(bucketID int64) ([]*BucketAccess, error) {
	rows, err := r.db.Query(
		`SELECT id, bucket_id, user_id, permission, created_at, updated_at
		FROM bucket_access
		WHERE bucket_id = ?
		ORDER BY created_at DESC`,
		bucketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accesses []*BucketAccess
	for rows.Next() {
		access := &BucketAccess{}
		if err := rows.Scan(
			&access.ID,
			&access.BucketID,
			&access.UserID,
			&access.Permission,
			&access.CreatedAt,
			&access.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accesses = append(accesses, access)
	}

	return accesses, rows.Err()
}

// ListByUser retrieves all access records for a user.
func (r *BucketAccessRepository) ListByUser(userID int64) ([]*BucketAccess, error) {
	rows, err := r.db.Query(
		`SELECT id, bucket_id, user_id, permission, created_at, updated_at
		FROM bucket_access
		WHERE user_id = ?
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accesses []*BucketAccess
	for rows.Next() {
		access := &BucketAccess{}
		if err := rows.Scan(
			&access.ID,
			&access.BucketID,
			&access.UserID,
			&access.Permission,
			&access.CreatedAt,
			&access.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accesses = append(accesses, access)
	}

	return accesses, rows.Err()
}

// Upsert inserts or updates a bucket access record.
func (r *BucketAccessRepository) Upsert(access *BucketAccess) error {
	existing, err := r.Get(access.BucketID, access.UserID)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.Create(access)
	}

	access.ID = existing.ID
	return r.Update(access)
}
