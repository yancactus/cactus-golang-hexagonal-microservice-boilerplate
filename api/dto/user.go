package dto

import "time"

// CreateUserReq represents the request to create a user
type CreateUserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserReq represents the request to update a user
type UpdateUserReq struct {
	Name string `json:"name" binding:"required"`
}

// GetUserReq represents the request to get a user
type GetUserReq struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// DeleteUserReq represents the request to delete a user
type DeleteUserReq struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// UserResp represents the user response
type UserResp struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
