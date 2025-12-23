package model

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// User domain errors are defined in domain_error.go

// User represents a user in the system
type User struct {
	ID        string
	Email     string
	Name      string
	Password  string // hashed password
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	events []DomainEvent
}

// NewUser creates a new user with validation
func NewUser(email, name, password string) (*User, error) {
	user := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	user.recordEvent(UserCreatedEvent{
		Email: email,
		Name:  name,
	})

	return user, nil
}

// Validate validates the user entity
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrUserEmailRequired
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return ErrUserEmailInvalid
	}

	if u.Name == "" {
		return ErrUserNameRequired
	}

	if u.Password == "" {
		return ErrUserPasswordRequired
	}

	return nil
}

// Update updates user information
func (u *User) Update(name string) error {
	if name == "" {
		return ErrUserNameRequired
	}

	u.Name = name
	u.UpdatedAt = time.Now()

	u.recordEvent(UserUpdatedEvent{
		ID:   u.ID,
		Name: name,
	})

	return nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(hashedPassword string) {
	u.Password = hashedPassword
	u.UpdatedAt = time.Now()
}

// MarkDeleted marks the user as deleted
func (u *User) MarkDeleted() {
	now := time.Now()
	u.DeletedAt = &now

	u.recordEvent(UserDeletedEvent{
		ID: u.ID,
	})
}

// Events returns and clears domain events
func (u *User) Events() []DomainEvent {
	events := u.events
	u.events = nil
	return events
}

func (u *User) recordEvent(event DomainEvent) {
	u.events = append(u.events, event)
}

// TableName returns the table name for GORM
func (u *User) TableName() string {
	return "users"
}

// User domain events
type UserCreatedEvent struct {
	Email string
	Name  string
}

func (e UserCreatedEvent) EventName() string { return "user.created" }

type UserUpdatedEvent struct {
	ID   string
	Name string
}

func (e UserUpdatedEvent) EventName() string { return "user.updated" }

type UserDeletedEvent struct {
	ID string
}

func (e UserDeletedEvent) EventName() string { return "user.deleted" }
