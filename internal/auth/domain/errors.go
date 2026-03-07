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

	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")

	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationCodeExpired = errors.New("verification code expired")
	ErrUserAlreadyActive       = errors.New("user already active")
)
