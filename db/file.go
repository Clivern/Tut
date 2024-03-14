// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// File represents a file stored in a bucket.
type File struct {
	ID          int64
	BucketID    int64
	Name        string
	Path        string
	ContentType string
	Size        int64
	ETag        string
	UserID      int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FileRepository handles database operations for files.
type FileRepository struct {
	db *sql.DB
}

// NewFileRepository creates a new file repository.
func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create inserts a new file into the database.
func (r *FileRepository) Create(file *File) error {
	result, err := r.db.Exec(
		`INSERT INTO files (bucket_id, name, path, content_type, size, etag, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		file.BucketID,
		file.Name,
		file.Path,
		file.ContentType,
		file.Size,
		file.ETag,
		file.UserID,
	)
	if err != nil {
		return err
	}

	file.ID, err = result.LastInsertId()
	return err
}

// GetByID retrieves a file by ID.
func (r *FileRepository) GetByID(id int64) (*File, error) {
	file := &File{}
	err := r.db.QueryRow(
		`SELECT id, bucket_id, name, path, content_type, size, etag, user_id, created_at, updated_at
		FROM files
		WHERE id = ?`,
		id,
	).Scan(
		&file.ID,
		&file.BucketID,
		&file.Name,
		&file.Path,
		&file.ContentType,
		&file.Size,
		&file.ETag,
		&file.UserID,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GetByName retrieves a file by name within a bucket.
func (r *FileRepository) GetByName(bucketID int64, name string) (*File, error) {
	file := &File{}
	err := r.db.QueryRow(
		`SELECT id, bucket_id, name, path, content_type, size, etag, user_id, created_at, updated_at
		FROM files
		WHERE bucket_id = ? AND name = ?`,
		bucketID,
		name,
	).Scan(
		&file.ID,
		&file.BucketID,
		&file.Name,
		&file.Path,
		&file.ContentType,
		&file.Size,
		&file.ETag,
		&file.UserID,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Update updates a file's information.
func (r *FileRepository) Update(file *File) error {
	_, err := r.db.Exec(
		`UPDATE files SET
			name = ?, path = ?, content_type = ?, size = ?, etag = ?, updated_at = ?
		WHERE id = ?`,
		file.Name,
		file.Path,
		file.ContentType,
		file.Size,
		file.ETag,
		time.Now().UTC(),
		file.ID,
	)
	return err
}

// Delete removes a file from the database.
func (r *FileRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM files WHERE id = ?", id)
	return err
}

// List retrieves all files in a bucket with pagination.
func (r *FileRepository) List(bucketID int64, limit, offset int) ([]*File, error) {
	rows, err := r.db.Query(
		`SELECT id, bucket_id, name, path, content_type, size, etag, user_id, created_at, updated_at
		FROM files
		WHERE bucket_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		bucketID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*File
	for rows.Next() {
		file := &File{}
		if err := rows.Scan(
			&file.ID,
			&file.BucketID,
			&file.Name,
			&file.Path,
			&file.ContentType,
			&file.Size,
			&file.ETag,
			&file.UserID,
			&file.CreatedAt,
			&file.UpdatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, rows.Err()
}

// Count returns the total number of files in a bucket.
func (r *FileRepository) Count(bucketID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM files WHERE bucket_id = ?", bucketID).Scan(&count)
	return count, err
}

// ListByPrefix retrieves files in a bucket matching a prefix.
func (r *FileRepository) ListByPrefix(bucketID int64, prefix string, limit, offset int) ([]*File, error) {
	rows, err := r.db.Query(
		`SELECT id, bucket_id, name, path, content_type, size, etag, user_id, created_at, updated_at
		FROM files
		WHERE bucket_id = ? AND name LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		bucketID,
		prefix+"%",
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*File
	for rows.Next() {
		file := &File{}
		if err := rows.Scan(
			&file.ID,
			&file.BucketID,
			&file.Name,
			&file.Path,
			&file.ContentType,
			&file.Size,
			&file.ETag,
			&file.UserID,
			&file.CreatedAt,
			&file.UpdatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, rows.Err()
}
