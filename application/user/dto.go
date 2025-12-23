package user

import "time"

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// Validate validates the create user input
func (i *CreateUserInput) Validate() error {
	if i.Email == "" {
		return ErrEmailRequired
	}
	if i.Name == "" {
		return ErrNameRequired
	}
	if i.Password == "" {
		return ErrPasswordRequired
	}
	if len(i.Password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	ID   string `json:"id" validate:"required,uuid"`
	Name string `json:"name" validate:"required"`
}

// Validate validates the update user input
func (i *UpdateUserInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	if i.Name == "" {
		return ErrNameRequired
	}
	return nil
}

// GetUserInput represents the input for getting a user
type GetUserInput struct {
	ID string `json:"id" validate:"required,uuid"`
}

// Validate validates the get user input
func (i *GetUserInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	return nil
}

// DeleteUserInput represents the input for deleting a user
type DeleteUserInput struct {
	ID string `json:"id" validate:"required,uuid"`
}

// Validate validates the delete user input
func (i *DeleteUserInput) Validate() error {
	if i.ID == "" {
		return ErrInvalidID
	}
	return nil
}

// ListUsersInput represents the input for listing users
type ListUsersInput struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// Validate validates the list users input
func (i *ListUsersInput) Validate() error {
	if i.Limit <= 0 {
		i.Limit = 10
	}
	if i.Offset < 0 {
		i.Offset = 0
	}
	return nil
}

// UserOutput represents the output for a user
type UserOutput struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsersOutput represents the output for listing users
type ListUsersOutput struct {
	Users []*UserOutput `json:"users"`
	Total int64         `json:"total"`
}
