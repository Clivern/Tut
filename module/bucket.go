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

// GetBucketByName retrieves a bucket by name.
func (b *Bucket) GetBucketByName(name string) (*db.Bucket, error) {
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
