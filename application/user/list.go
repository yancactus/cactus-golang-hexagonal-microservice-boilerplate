package user

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// ListUsersUseCase handles listing users
type ListUsersUseCase struct {
	userService service.IUserService
}

// NewListUsersUseCase creates a new ListUsersUseCase
func NewListUsersUseCase(userService service.IUserService) *ListUsersUseCase {
	return &ListUsersUseCase{
		userService: userService,
	}
}

// Execute lists users with pagination
func (uc *ListUsersUseCase) Execute(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	users, total, err := uc.userService.List(ctx, input.Offset, input.Limit)
	if err != nil {
		return nil, err
	}

	output := &ListUsersOutput{
		Users: make([]*UserOutput, len(users)),
		Total: total,
	}

	for i, u := range users {
		output.Users[i] = &UserOutput{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	return output, nil
}
