package user

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// UpdateUserUseCase handles updating a user
type UpdateUserUseCase struct {
	userService service.IUserService
}

// NewUpdateUserUseCase creates a new UpdateUserUseCase
func NewUpdateUserUseCase(userService service.IUserService) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userService: userService,
	}
}

// Execute updates a user
func (uc *UpdateUserUseCase) Execute(ctx context.Context, input *UpdateUserInput) (*UserOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	user, err := uc.userService.Update(ctx, input.ID, input.Name)
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
