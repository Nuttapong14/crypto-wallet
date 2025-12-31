package repositories

import "errors"

var (
	// ErrNotFound indicates that the requested entity does not exist.
	ErrNotFound = errors.New("repository: entity not found")
	// ErrDuplicate indicates that the entity being created already exists.
	ErrDuplicate = errors.New("repository: duplicate entity")
)
