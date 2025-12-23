package user

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// DeleteUserUseCase handles deleting a user
type DeleteUserUseCase struct {
	userService service.IUserService
}

// NewDeleteUserUseCase creates a new DeleteUserUseCase
func NewDeleteUserUseCase(userService service.IUserService) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userService: userService,
	}
}

// Execute deletes a user
func (uc *DeleteUserUseCase) Execute(ctx context.Context, input *DeleteUserInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	return uc.userService.Delete(ctx, input.ID)
}
