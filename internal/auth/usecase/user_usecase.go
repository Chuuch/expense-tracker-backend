package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces"
	"github.com/chuuch/expense-tracker-backend/utils"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound            = domain.ErrUserNotFound
	ErrInvalidCredentials      = domain.ErrInvalidCredentials
	ErrUserAlreadyExists       = domain.ErrUserAlreadyExists
	ErrUserAlreadyDeleted      = domain.ErrUserAlreadyDeleted
	ErrUserNotDeleted          = domain.ErrUserNotDeleted
	ErrUserNotActive           = domain.ErrUserNotActive
	ErrUserNotLocked           = domain.ErrUserNotLocked
	ErrUserNotPending          = domain.ErrUserNotPending
	ErrUserNotAdmin            = domain.ErrUserNotAdmin
	ErrUserNotSupport          = domain.ErrUserNotSupport
	ErrInvalidVerificationCode = domain.ErrInvalidVerificationCode
	ErrVerificationCodeExpired = domain.ErrVerificationCodeExpired
	ErrUserAlreadyActive       = domain.ErrUserAlreadyActive
)

type UserUsecase struct {
	userRepo             interfaces.UserRepository
	verificationEnqueuer interfaces.VerificationEmailEnqueuer
}

func NewUserUsecase(userRepo interfaces.UserRepository, verificationEnqueuer interfaces.VerificationEmailEnqueuer) *UserUsecase {
	return &UserUsecase{
		userRepo:             userRepo,
		verificationEnqueuer: verificationEnqueuer,
	}
}

func (u *UserUsecase) Register(
	ctx context.Context,
	email string,
	password string,
	username string,
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
		username,
		phone,
		address,
		city,
		state,
		zip,
		country,
	)

	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return nil, fmt.Errorf("usecase.Register: generate verification code: %w", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	user.VerificationCode = &code
	user.VerificationCodeExpiresAt = &expiresAt

	createdUser, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("usecase.CreateUser: %w", err)
	}

	if u.verificationEnqueuer != nil {
		_ = u.verificationEnqueuer.EnqueueSendVerificationEmail(ctx, createdUser.ID)
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

	if user.Status != domain.StatusActive {
		return nil, ErrUserNotActive
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := u.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("usecase.UpdateLastLoginAt: %w", err)
	}

	return user, nil
}

func (u *UserUsecase) VerifyEmail(ctx context.Context, email, code string) (*domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("usecase.VerifyEmail: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.VerificationCode == nil || user.VerificationCodeExpiresAt == nil {
		return nil, ErrInvalidVerificationCode
	}
	if time.Now().After(*user.VerificationCodeExpiresAt) {
		return nil, ErrVerificationCodeExpired
	}
	if *user.VerificationCode != code {
		return nil, ErrInvalidVerificationCode
	}
	user.VerificationCode = nil
	user.VerificationCodeExpiresAt = nil
	user.Status = domain.StatusActive
	user.UpdatedAt = time.Now()
	if err := u.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("usecase.VerifyEmail UpdateUser: %w", err)
	}
	return user, nil
}

func (u *UserUsecase) ResendVerificationEmail(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("usecase.ResendVerificationEmail: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}
	if user.Status != domain.StatusPending {
		return ErrUserAlreadyActive
	}
	code, err := utils.GenerateVerificationCode()
	if err != nil {
		return fmt.Errorf("usecase.ResendVerificationEmail: generate code: %w", err)
	}
	expiresAt := time.Now().Add(15 * time.Minute)
	user.VerificationCode = &code
	user.VerificationCodeExpiresAt = &expiresAt
	user.UpdatedAt = time.Now()
	if err := u.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("usecase.ResendVerificationEmail UpdateUser: %w", err)
	}
	if u.verificationEnqueuer != nil {
		_ = u.verificationEnqueuer.EnqueueSendVerificationEmail(ctx, user.ID)
	}
	return nil
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
	username string,
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
		Username: username,
		Phone:    phone,
		Address:  address,
		City:     city,
		State:    state,
		Zip:      zip,
		Country:  country,
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
