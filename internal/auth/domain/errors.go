package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyDeleted = errors.New("user already deleted")
	ErrUserNotDeleted     = errors.New("user not deleted")
	ErrUserNotActive      = errors.New("user not active")
	ErrUserNotLocked      = errors.New("user not locked")
	ErrUserNotPending     = errors.New("user not pending")
	ErrUserNotAdmin       = errors.New("user not admin")
	ErrUserNotSupport     = errors.New("user not support")
)
