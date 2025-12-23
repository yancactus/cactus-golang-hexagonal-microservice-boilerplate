package user

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// CreateUserUseCase handles user creation
type CreateUserUseCase struct {
	userService service.IUserService
}

// NewCreateUserUseCase creates a new CreateUserUseCase
func NewCreateUserUseCase(userService service.IUserService) *CreateUserUseCase {
	return &CreateUserUseCase{
		userService: userService,
	}
}

// Execute creates a new user
func (uc *CreateUserUseCase) Execute(ctx context.Context, input *CreateUserInput) (*UserOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	user, err := uc.userService.Create(ctx, input.Email, input.Name, input.Password)
	if err != nil {
		return nil, err
	}

	return &UserOutput{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
