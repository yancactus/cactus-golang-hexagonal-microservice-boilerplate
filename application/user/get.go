package user

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// GetUserUseCase handles getting a user by ID
type GetUserUseCase struct {
	userService service.IUserService
}

// NewGetUserUseCase creates a new GetUserUseCase
func NewGetUserUseCase(userService service.IUserService) *GetUserUseCase {
	return &GetUserUseCase{
		userService: userService,
	}
}

// Execute retrieves a user by ID
func (uc *GetUserUseCase) Execute(ctx context.Context, input *GetUserInput) (*UserOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	user, err := uc.userService.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return &UserOutput{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
