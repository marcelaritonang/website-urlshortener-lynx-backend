package types

import "errors"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// URL related errors
var (
	ErrShortCodeTaken    = errors.New("short code is already taken")
	ErrInvalidShortCode  = errors.New("short code can only contain letters, numbers, hyphens, and underscores")
	ErrGenerateShortCode = errors.New("failed to generate unique short code")
	ErrURLNotFound       = errors.New("url not found")
	ErrInvalidURLID      = errors.New("invalid url id")
	ErrUnauthorized      = errors.New("unauthorized access")
)

var (
	// Auth errors
	ErrMissingToken         = errors.New("authorization header required")
	ErrExpiredToken         = errors.New("token has expired")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrInvalidClaims        = errors.New("invalid token claims")
	ErrInvalidUserID        = errors.New("invalid user ID in token")
	ErrInvalidUUID          = errors.New("invalid UUID format")
)

// User related errors
var (
	ErrUserExists                 = errors.New("user already exists")
	ErrUserNotFound               = errors.New("user not found")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrInvalidToken               = errors.New("invalid token")
	ErrTokenExpired               = errors.New("token has expired")
	ErrPasswordMismatch           = errors.New("password does not match")
	ErrInvalidOrExpiredResetToken = errors.New("invalid or expired reset token")
	ErrResetTokenHasExpired       = errors.New("reset token has expired")
)

// Generic errors
var (
	ErrInvalidInput     = errors.New("invalid input data")
	ErrDatabaseError    = errors.New("database error occurred")
	ErrCacheError       = errors.New("cache error occurred")
	ErrInternalError    = errors.New("internal server error")
	ErrResourceNotFound = errors.New("resource not found")
)
