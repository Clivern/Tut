// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Bucket represents a bucket in the database.
type Bucket struct {
	ID           int64
	Name         string
	Region       string
	Tags         map[string]string
	Versioning   bool
	PublicAccess bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	tagsJSON, err := json.Marshal(bucket.Tags)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(
		`INSERT INTO buckets (name, region, tags, versioning, public_access)
		VALUES (?, ?, ?, ?, ?)`,
		bucket.Name,
		bucket.Region,
		string(tagsJSON),
		bucket.Versioning,
		bucket.PublicAccess,
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
	var tagsJSON string
	err := r.db.QueryRow(
		`SELECT id, name, region, tags, versioning, public_access, created_at, updated_at
		FROM buckets
		WHERE id = ?`,
		id,
	).Scan(
		&bucket.ID,
		&bucket.Name,
		&bucket.Region,
		&tagsJSON,
		&bucket.Versioning,
		&bucket.PublicAccess,
		&bucket.CreatedAt,
		&bucket.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &bucket.Tags); err != nil {
		bucket.Tags = make(map[string]string)
	}

	return bucket, nil
}

// GetByName retrieves a bucket by name.
func (r *BucketRepository) GetByName(name string) (*Bucket, error) {
	bucket := &Bucket{}
	var tagsJSON string
	err := r.db.QueryRow(
		`SELECT id, name, region, tags, versioning, public_access, created_at, updated_at
		FROM buckets
		WHERE name = ?`,
		name,
	).Scan(
		&bucket.ID,
		&bucket.Name,
		&bucket.Region,
		&tagsJSON,
		&bucket.Versioning,
		&bucket.PublicAccess,
		&bucket.CreatedAt,
		&bucket.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &bucket.Tags); err != nil {
		bucket.Tags = make(map[string]string)
	}

	return bucket, nil
}

// Update updates a bucket's information.
func (r *BucketRepository) Update(bucket *Bucket) error {
	tagsJSON, err := json.Marshal(bucket.Tags)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`UPDATE buckets SET
			name = ?, region = ?, tags = ?, versioning = ?, public_access = ?, updated_at = ?
		WHERE id = ?`,
		bucket.Name,
		bucket.Region,
		string(tagsJSON),
		bucket.Versioning,
		bucket.PublicAccess,
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

// List retrieves all buckets with pagination.
func (r *BucketRepository) List(limit, offset int) ([]*Bucket, error) {
	rows, err := r.db.Query(
		`SELECT id, name, region, tags, versioning, public_access, created_at, updated_at
		FROM buckets
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
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
		var tagsJSON string
		if err := rows.Scan(
			&bucket.ID,
			&bucket.Name,
			&bucket.Region,
			&tagsJSON,
			&bucket.Versioning,
			&bucket.PublicAccess,
			&bucket.CreatedAt,
			&bucket.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsJSON), &bucket.Tags); err != nil {
			bucket.Tags = make(map[string]string)
		}

		buckets = append(buckets, bucket)
	}

	return buckets, rows.Err()
}

// Count returns the total number of buckets.
func (r *BucketRepository) Count() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM buckets").Scan(&count)
	return count, err
}

// BucketMeta represents metadata associated with a bucket.
type BucketMeta struct {
	ID        int64
	Key       string
	Value     string
	BucketID  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BucketMetaRepository handles database operations for bucket metadata.
type BucketMetaRepository struct {
	db *sql.DB
}

// NewBucketMetaRepository creates a new bucket meta repository.
func NewBucketMetaRepository(db *sql.DB) *BucketMetaRepository {
	return &BucketMetaRepository{db: db}
}

// Create inserts new metadata for a bucket.
func (r *BucketMetaRepository) Create(bucketID int64, key, value string) error {
	_, err := r.db.Exec(
		"INSERT INTO buckets_meta (bucket_id, key, value) VALUES (?, ?, ?)",
		bucketID,
		key,
		value,
	)
	return err
}

// Get retrieves metadata for a bucket by key.
func (r *BucketMetaRepository) Get(bucketID int64, key string) (*BucketMeta, error) {
	meta := &BucketMeta{}
	err := r.db.QueryRow(
		`SELECT id, key, value, bucket_id, created_at, updated_at
		FROM buckets_meta
		WHERE bucket_id = ? AND key = ?`,
		bucketID,
		key,
	).Scan(
		&meta.ID,
		&meta.Key,
		&meta.Value,
		&meta.BucketID,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// Update updates metadata for a bucket.
func (r *BucketMetaRepository) Update(bucketID int64, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE buckets_meta SET
			value = ?, updated_at = ?
		WHERE bucket_id = ? AND key = ?`,
		value,
		time.Now().UTC(),
		bucketID,
		key,
	)
	return err
}

// Delete removes metadata for a bucket.
func (r *BucketMetaRepository) Delete(bucketID int64, key string) error {
	_, err := r.db.Exec(
		"DELETE FROM buckets_meta WHERE bucket_id = ? AND key = ?",
		bucketID,
		key,
	)
	return err
}

// ListByBucket retrieves all metadata for a bucket.
func (r *BucketMetaRepository) ListByBucket(bucketID int64) ([]*BucketMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, key, value, bucket_id, created_at, updated_at
		FROM buckets_meta
		WHERE bucket_id = ?
		ORDER BY key`,
		bucketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metadata []*BucketMeta
	for rows.Next() {
		meta := &BucketMeta{}
		if err := rows.Scan(
			&meta.ID,
			&meta.Key,
			&meta.Value,
			&meta.BucketID,
			&meta.CreatedAt,
			&meta.UpdatedAt,
		); err != nil {
			return nil, err
		}
		metadata = append(metadata, meta)
	}

	return metadata, rows.Err()
}

// Upsert inserts or updates metadata for a bucket.
func (r *BucketMetaRepository) Upsert(bucketID int64, key, value string) error {
	existing, err := r.Get(bucketID, key)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.Create(bucketID, key, value)
	}

	return r.Update(bucketID, key, value)
}
