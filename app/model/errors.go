package model

import "errors"

var (
	ErrNotFound     = errors.New("data not found")
	ErrDuplicate    = errors.New("data already exists")
	ErrForeignKey   = errors.New("invalid reference or data still in use")
	ErrConstraint   = errors.New("data violates a constraint")
	ErrBusinessRule = errors.New("business rule violation")
	ErrInvalidState = errors.New("invalid state for this operation")
)
