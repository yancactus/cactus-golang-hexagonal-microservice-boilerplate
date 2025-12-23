package user

import "errors"

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrNameRequired     = errors.New("name is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrInvalidID        = errors.New("invalid user ID")
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailTaken       = errors.New("email is already taken")
)
