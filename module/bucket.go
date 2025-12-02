// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"strings"

	"github.com/clivern/tut/db"
)

// Bucket module errors
var (
	ErrBucketNotFound          = errors.New("bucket not found")
	ErrBucketNameAlreadyExists = errors.New("bucket with this name already exists")
	ErrInvalidBucketName       = errors.New("invalid bucket name")
)

// Bucket handles bucket management operations.
type Bucket struct {
	BucketRepository *db.BucketRepository
}

// NewBucket creates a new bucket module instance.
func NewBucket(repo *db.BucketRepository) *Bucket {
	return &Bucket{BucketRepository: repo}
}

// CreateBucketOptions contains options for creating a bucket.
type CreateBucketOptions struct {
	Name         string
	Region       string
	Tags         map[string]string
	Versioning   bool
	PublicAccess bool
}

// CreateBucket creates a new bucket.
func (b *Bucket) CreateBucket(options *CreateBucketOptions) (*db.Bucket, error) {
	// Validate bucket name
	if err := validateBucketName(options.Name); err != nil {
		return nil, err
	}

	existingBucket, err := b.BucketRepository.GetByName(options.Name)
	if err != nil {
		return nil, err
	}
	if existingBucket != nil {
		return nil, ErrBucketNameAlreadyExists
	}

	bucket := &db.Bucket{
		Name:         strings.ToLower(strings.TrimSpace(options.Name)),
		Region:       strings.TrimSpace(options.Region),
		Tags:         options.Tags,
		Versioning:   options.Versioning,
		PublicAccess: options.PublicAccess,
	}

	if bucket.Tags == nil {
		bucket.Tags = make(map[string]string)
	}

	if err := b.BucketRepository.Create(bucket); err != nil {
		return nil, err
	}

	return bucket, nil
}

// GetBucket retrieves a bucket by name.
func (b *Bucket) GetBucket(name string) (*db.Bucket, error) {
	bucket, err := b.BucketRepository.GetByName(name)
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, ErrBucketNotFound
	}
	return bucket, nil
}

// GetBucketByID retrieves a bucket by ID.
func (b *Bucket) GetBucketByID(id int64) (*db.Bucket, error) {
	bucket, err := b.BucketRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, ErrBucketNotFound
	}
	return bucket, nil
}

// UpdateBucketOptions contains options for updating a bucket.
type UpdateBucketOptions struct {
	Name         string
	Region       string
	Tags         map[string]string
	Versioning   *bool
	PublicAccess *bool
}

// UpdateBucket updates an existing bucket.
func (b *Bucket) UpdateBucket(name string, options *UpdateBucketOptions) (*db.Bucket, error) {
	bucket, err := b.BucketRepository.GetByName(name)
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, ErrBucketNotFound
	}

	if options.Name != "" && options.Name != bucket.Name {
		if err := validateBucketName(options.Name); err != nil {
			return nil, err
		}
		existingBucket, err := b.BucketRepository.GetByName(options.Name)
		if err != nil {
			return nil, err
		}
		if existingBucket != nil && existingBucket.ID != bucket.ID {
			return nil, ErrBucketNameAlreadyExists
		}
		bucket.Name = strings.ToLower(strings.TrimSpace(options.Name))
	}

	if options.Region != "" {
		bucket.Region = strings.TrimSpace(options.Region)
	}

	if options.Tags != nil {
		bucket.Tags = options.Tags
	}

	if options.Versioning != nil {
		bucket.Versioning = *options.Versioning
	}

	if options.PublicAccess != nil {
		bucket.PublicAccess = *options.PublicAccess
	}

	if err := b.BucketRepository.Update(bucket); err != nil {
		return nil, err
	}

	return bucket, nil
}

// UpdateBucketByID updates an existing bucket by ID.
func (b *Bucket) UpdateBucketByID(id int64, options *UpdateBucketOptions) (*db.Bucket, error) {
	bucket, err := b.BucketRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, ErrBucketNotFound
	}

	if options.Name != "" && options.Name != bucket.Name {
		if err := validateBucketName(options.Name); err != nil {
			return nil, err
		}
		existingBucket, err := b.BucketRepository.GetByName(options.Name)
		if err != nil {
			return nil, err
		}
		if existingBucket != nil && existingBucket.ID != bucket.ID {
			return nil, ErrBucketNameAlreadyExists
		}
		bucket.Name = strings.ToLower(strings.TrimSpace(options.Name))
	}

	if options.Region != "" {
		bucket.Region = strings.TrimSpace(options.Region)
	}

	if options.Tags != nil {
		bucket.Tags = options.Tags
	}

	if options.Versioning != nil {
		bucket.Versioning = *options.Versioning
	}

	if options.PublicAccess != nil {
		bucket.PublicAccess = *options.PublicAccess
	}

	if err := b.BucketRepository.Update(bucket); err != nil {
		return nil, err
	}

	return bucket, nil
}

// ListBucketsOptions contains options for listing buckets.
type ListBucketsOptions struct {
	Limit  int
	Offset int
}

// ListBucketsResult contains the result of listing buckets.
type ListBucketsResult struct {
	Buckets []*db.Bucket
	Total   int64
}

// ListBuckets retrieves a list of buckets with pagination.
func (b *Bucket) ListBuckets(options *ListBucketsOptions) (*ListBucketsResult, error) {
	buckets, err := b.BucketRepository.List(options.Limit, options.Offset)
	if err != nil {
		return nil, err
	}

	total, err := b.BucketRepository.Count()
	if err != nil {
		return nil, err
	}

	return &ListBucketsResult{
		Buckets: buckets,
		Total:   total,
	}, nil
}

// DeleteBucketByID deletes a bucket by ID.
func (b *Bucket) DeleteBucketByID(id int64) error {
	bucket, err := b.BucketRepository.GetByID(id)
	if err != nil {
		return err
	}
	if bucket == nil {
		return ErrBucketNotFound
	}

	return b.BucketRepository.Delete(bucket.ID)
}

// validateBucketName validates a bucket name according to common S3-compatible rules.
func validateBucketName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidBucketName
	}

	// Bucket names must be 3-63 characters long
	if len(name) < 3 || len(name) > 63 {
		return ErrInvalidBucketName
	}

	// Bucket names must be lowercase
	if name != strings.ToLower(name) {
		return ErrInvalidBucketName
	}

	// Bucket names can only contain lowercase letters, numbers, dots, and hyphens
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.' ||
			char == '-') {
			return ErrInvalidBucketName
		}
	}

	// Bucket names cannot start or end with a dot or hyphen
	if name[0] == '.' || name[0] == '-' ||
		name[len(name)-1] == '.' || name[len(name)-1] == '-' {
		return ErrInvalidBucketName
	}

	// Bucket names cannot contain consecutive dots
	if strings.Contains(name, "..") {
		return ErrInvalidBucketName
	}

	return nil
}
