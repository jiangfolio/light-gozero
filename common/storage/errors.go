package storage

import "errors"

var (
	ErrInvalidKey          = errors.New("storage: invalid key")
	ErrInvalidPart         = errors.New("storage: invalid part")
	ErrChecksumMismatch    = errors.New("storage: checksum mismatch")
	ErrUploadNotFound      = errors.New("storage: upload not found")
	ErrUploadExpired       = errors.New("storage: upload expired")
	ErrUploadState         = errors.New("storage: invalid upload state")
	ErrTooLarge            = errors.New("storage: object or part too large")
	ErrPartExistsConflict  = errors.New("storage: part exists with different checksum")
	ErrDirectUploadOnly    = errors.New("storage: direct upload only")
	ErrNotSupported        = errors.New("storage: not supported")
	ErrObjectNotFound      = errors.New("storage: object not found")
	ErrInsufficientSpace   = errors.New("storage: insufficient disk space")
	ErrInvalidConfig       = errors.New("storage: invalid config")
	ErrMerkleMismatch      = errors.New("storage: merkle mismatch")
)
