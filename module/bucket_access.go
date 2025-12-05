// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package module

import (
	"errors"

	"github.com/clivern/tut/db"
)

// BucketAccess module errors
var (
	ErrBucketAccessNotFound      = errors.New("bucket access not found")
	ErrInvalidPermission         = errors.New("invalid permission")
	ErrBucketAccessAlreadyExists = errors.New("bucket access already exists")
)

// BucketAccess handles bucket access management operations.
type BucketAccess struct {
	BucketAccessRepository *db.BucketAccessRepository
	BucketRepository       *db.BucketRepository
	UserRepository         *db.UserRepository
}

// NewBucketAccess creates a new bucket access module instance.
func NewBucketAccess(
	bucketAccessRepo *db.BucketAccessRepository,
	bucketRepo *db.BucketRepository,
	userRepo *db.UserRepository,
) *BucketAccess {
	return &BucketAccess{
		BucketAccessRepository: bucketAccessRepo,
		BucketRepository:       bucketRepo,
		UserRepository:         userRepo,
	}
}

// ValidatePermission validates that a permission is one of the allowed values.
func ValidatePermission(permission string) bool {
	return permission == db.BucketPermissionRead ||
		permission == db.BucketPermissionWrite ||
		permission == db.BucketPermissionAdmin
}

// CreateBucketAccessOptions contains options for creating bucket access.
type CreateBucketAccessOptions struct {
	BucketID   int64
	UserID     int64
	Permission string
}

// CreateBucketAccess creates a new bucket access record.
func (ba *BucketAccess) CreateBucketAccess(options *CreateBucketAccessOptions) (*db.BucketAccess, error) {
	// Validate permission
	if !ValidatePermission(options.Permission) {
		return nil, ErrInvalidPermission
	}

	// Verify bucket exists
	bucket, err := ba.BucketRepository.GetByID(options.BucketID)
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, ErrBucketNotFound
	}

	// Verify user exists
	user, err := ba.UserRepository.GetByID(options.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if access already exists
	existing, err := ba.BucketAccessRepository.Get(options.BucketID, options.UserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrBucketAccessAlreadyExists
	}

	access := &db.BucketAccess{
		BucketID:   options.BucketID,
		UserID:     options.UserID,
		Permission: options.Permission,
	}

	if err := ba.BucketAccessRepository.Create(access); err != nil {
		return nil, err
	}

	return access, nil
}

// GetBucketAccess retrieves bucket access by bucket ID and user ID.
func (ba *BucketAccess) GetBucketAccess(bucketID, userID int64) (*db.BucketAccess, error) {
	access, err := ba.BucketAccessRepository.Get(bucketID, userID)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, ErrBucketAccessNotFound
	}
	return access, nil
}

// GetBucketAccessByID retrieves bucket access by ID.
func (ba *BucketAccess) GetBucketAccessByID(id int64) (*db.BucketAccess, error) {
	access, err := ba.BucketAccessRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, ErrBucketAccessNotFound
	}
	return access, nil
}

// UpdateBucketAccessOptions contains options for updating bucket access.
type UpdateBucketAccessOptions struct {
	Permission string
}

// UpdateBucketAccess updates an existing bucket access record.
func (ba *BucketAccess) UpdateBucketAccess(bucketID, userID int64, options *UpdateBucketAccessOptions) (*db.BucketAccess, error) {
	// Validate permission
	if !ValidatePermission(options.Permission) {
		return nil, ErrInvalidPermission
	}

	// Get existing access
	access, err := ba.BucketAccessRepository.Get(bucketID, userID)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, ErrBucketAccessNotFound
	}

	access.Permission = options.Permission

	if err := ba.BucketAccessRepository.Update(access); err != nil {
		return nil, err
	}

	return access, nil
}

// UpdateBucketAccessByID updates an existing bucket access record by ID.
func (ba *BucketAccess) UpdateBucketAccessByID(id int64, options *UpdateBucketAccessOptions) (*db.BucketAccess, error) {
	// Validate permission
	if !ValidatePermission(options.Permission) {
		return nil, ErrInvalidPermission
	}

	// Get existing access
	access, err := ba.BucketAccessRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, ErrBucketAccessNotFound
	}

	access.Permission = options.Permission

	if err := ba.BucketAccessRepository.Update(access); err != nil {
		return nil, err
	}

	return access, nil
}

// ListBucketAccessOptions contains options for listing bucket access.
type ListBucketAccessOptions struct {
	BucketID int64
}

// ListBucketAccessResult contains the result of listing bucket access.
type ListBucketAccessResult struct {
	Accesses []*db.BucketAccess
}

// ListBucketAccess retrieves all access records for a bucket.
func (ba *BucketAccess) ListBucketAccess(options *ListBucketAccessOptions) (*ListBucketAccessResult, error) {
	// Verify bucket exists
	bucket, err := ba.BucketRepository.GetByID(options.BucketID)
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, ErrBucketNotFound
	}

	accesses, err := ba.BucketAccessRepository.ListByBucket(options.BucketID)
	if err != nil {
		return nil, err
	}

	return &ListBucketAccessResult{
		Accesses: accesses,
	}, nil
}

// DeleteBucketAccess removes bucket access by bucket ID and user ID.
func (ba *BucketAccess) DeleteBucketAccess(bucketID, userID int64) error {
	// Check if access exists
	access, err := ba.BucketAccessRepository.Get(bucketID, userID)
	if err != nil {
		return err
	}
	if access == nil {
		return ErrBucketAccessNotFound
	}

	return ba.BucketAccessRepository.Delete(bucketID, userID)
}

// DeleteBucketAccessByID removes bucket access by ID.
func (ba *BucketAccess) DeleteBucketAccessByID(id int64) error {
	// Check if access exists
	access, err := ba.BucketAccessRepository.GetByID(id)
	if err != nil {
		return err
	}
	if access == nil {
		return ErrBucketAccessNotFound
	}

	return ba.BucketAccessRepository.DeleteByID(id)
}
