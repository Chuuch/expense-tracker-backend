package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/ports"
	"github.com/chuuch/expense-tracker-backend/utils"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type UserUsecase struct {
	userRepo ports.UserRepository
}

func NewUserUsecase(userRepo ports.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (u *UserUsecase) Register(
	ctx context.Context,
	email string,
	password string,
	firstName string,
	lastName string,
	phone string,
	address string,
	city string,
	state string,
	zip string,
	country string,
) (*domain.User, error) {
	existing, _ := u.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	id := utils.GenerateULID()

	user := domain.NewUser(
		id,
		email,
		string(hashed),
		firstName,
		lastName,
		phone,
		address,
		city,
		state,
		zip,
		country,
	)

	createdUser, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("usecase.CreateUser: %w", err)
	}

	return createdUser, nil
}

func (u *UserUsecase) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := u.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := u.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("usecase.UpdateLastLoginAt: %w", err)
	}

	return user, nil
}

func (u *UserUsecase) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetByEmail: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (u *UserUsecase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetByID: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (u *UserUsecase) UpdateUser(
	ctx context.Context,
	id string,
	firstName string,
	lastName string,
	phone string,
	address string,
	city string,
	state string,
	zip string,
	country string,
) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, fmt.Errorf("usecase.GetByID: %w", err)
	}

	user.Profile = domain.Profile{
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Address:   address,
		City:      city,
		State:     state,
		Zip:       zip,
		Country:   country,
	}

	user.UpdatedAt = time.Now()

	if err := u.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("usecase.UpdateUser: %w", err)
	}

	return user, nil
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id string) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("usecase.GetByID: %w", err)
	}

	if user == nil {
		return ErrUserNotFound
	}

	if err := u.userRepo.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("usecase.DeleteUser: %w", err)
	}

	return nil
}
