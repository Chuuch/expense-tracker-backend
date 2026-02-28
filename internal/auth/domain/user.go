package domain

import "time"

type UserStatus string

const (
	StatusPending UserStatus = "pending"
	StatusActive  UserStatus = "active"
	StatusLocked  UserStatus = "locked"
)

type UserRole string

const (
	RoleUser    UserRole = "user"
	RoleAdmin   UserRole = "admin"
	RoleSupport UserRole = "support"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	IsMFAEnabled bool
	Role         UserRole
	Status       UserStatus
	Profile      Profile
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
	DeletedAt    *time.Time
}

type Profile struct {
	FirstName string
	LastName  string
	Phone     string
	Address   string
	City      string
	State     string
	Zip       string
	Country   string
}

func NewUser(
	id,
	email,
	passwordHash,
	firstName,
	lastName,
	phone,
	address,
	city,
	state,
	zip,
	country string) *User {
	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleUser,
		Status:       StatusPending,
		Profile: Profile{
			FirstName: firstName,
			LastName:  lastName,
			Phone:     phone,
			Address:   address,
			City:      city,
			State:     state,
			Zip:       zip,
			Country:   country,
		},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastLoginAt: nil,
		DeletedAt:   nil,
	}
}
