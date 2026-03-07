package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase"
	"github.com/chuuch/expense-tracker-backend/internal/auth/usecase/interfaces/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_Register_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(&domain.User{ID: "existing-id"}, nil).
		Times(1)

	_, err := uc.Register(ctx, "test@example.com", "Password123!", "Test", "+15550001111", "", "", "", "", "")
	if !errors.Is(err, usecase.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestUserUsecase_Register_Success_HashesPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(nil, nil).
		Times(1)

	repo.EXPECT().
		CreateUser(ctx, gomock.AssignableToTypeOf(&domain.User{})).
		DoAndReturn(func(_ context.Context, u *domain.User) (*domain.User, error) {
			if u.ID == "" {
				t.Fatalf("expected generated user ID")
			}
			if u.PasswordHash == "Password123!" {
				t.Fatalf("password should be hashed")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("Password123!")); err != nil {
				t.Fatalf("hash does not match original password: %v", err)
			}
			return u, nil
		}).
		Times(1)

	got, err := uc.Register(ctx, "test@example.com", "Password123!", "Test", "+15550001111", "", "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected created user, got nil")
	}
}

func TestUserUsecase_Login_UserNotFound_ReturnsInvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByEmail(ctx, "missing@example.com").
		Return(nil, nil).
		Times(1)

	_, err := uc.Login(ctx, "missing@example.com", "Password123!")
	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestUserUsecase_Login_WrongPassword_ReturnsInvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate hash: %v", err)
	}

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(&domain.User{
			ID:           "u1",
			Email:        "test@example.com",
			PasswordHash: string(hash),
		}, nil).
		Times(1)

	_, err = uc.Login(ctx, "test@example.com", "WrongPassword123!")
	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestUserUsecase_Login_Success_UpdatesLastLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate hash: %v", err)
	}

	user := &domain.User{
		ID:           "u1",
		Email:        "test@example.com",
		PasswordHash: string(hash),
	}

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(user, nil).
		Times(1)

	repo.EXPECT().
		UpdateUser(ctx, gomock.AssignableToTypeOf(&domain.User{})).
		DoAndReturn(func(_ context.Context, u *domain.User) error {
			if u.LastLoginAt == nil {
				t.Fatalf("expected LastLoginAt to be set")
			}
			return nil
		}).
		Times(1)

	got, err := uc.Login(ctx, "test@example.com", "Password123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.LastLoginAt == nil {
		t.Fatalf("expected returned user with LastLoginAt set")
	}
}

func TestUserUsecase_GetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByID(ctx, "missing-id").
		Return(nil, nil).
		Times(1)

	_, err := uc.GetByID(ctx, "missing-id")
	if !errors.Is(err, usecase.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestUserUsecase_UpdateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	existing := &domain.User{
		ID:           "u1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		Role:         domain.RoleUser,
		Status:       domain.StatusPending,
		Profile: domain.Profile{
			Username: "Old",
			Phone:    "123",
			Address:  "Old Addr",
			City:     "Old City",
			State:    "OS",
			Zip:      "00000",
			Country:  "US",
		},
	}

	repo.EXPECT().
		GetByID(ctx, "u1").
		Return(existing, nil).
		Times(1)

	repo.EXPECT().
		UpdateUser(ctx, gomock.AssignableToTypeOf(&domain.User{})).
		DoAndReturn(func(_ context.Context, u *domain.User) error {
			if u.Profile.Username != "New" {
				t.Fatalf("profile not updated correctly: %+v", u.Profile)
			}
			if u.UpdatedAt.IsZero() {
				t.Fatalf("expected UpdatedAt to be set")
			}
			return nil
		}).
		Times(1)

	got, err := uc.UpdateUser(ctx, "u1", "New", "+15550001111", "123 Main", "Austin", "TX", "78701", "US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected updated user, got nil")
	}
}

func TestUserUsecase_UpdateUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByID(ctx, "missing-id").
		Return(nil, nil).
		Times(1)

	_, err := uc.UpdateUser(ctx, "missing-id", "A", "123", "", "", "", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserUsecase_DeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByID(ctx, "u1").
		Return(&domain.User{ID: "u1"}, nil).
		Times(1)

	repo.EXPECT().
		DeleteUser(ctx, "u1").
		Return(nil).
		Times(1)

	err := uc.DeleteUser(ctx, "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserUsecase_DeleteUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByID(ctx, "missing-id").
		Return(nil, nil).
		Times(1)

	err := uc.DeleteUser(ctx, "missing-id")
	if !errors.Is(err, usecase.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestUserUsecase_GetByEmail_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(nil, errors.New("db down")).
		Times(1)

	_, err := uc.GetByEmail(ctx, "test@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserUsecase_VerifyEmail_UserNotFound_ReturnsErrUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	repo.EXPECT().
		GetByEmail(ctx, "missing@example.com").
		Return(nil, nil).
		Times(1)

	_, err := uc.VerifyEmail(ctx, "missing@example.com", "123456")
	if !errors.Is(err, usecase.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestUserUsecase_VerifyEmail_NoVerificationCode_ReturnsErrInvalidVerificationCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	user := &domain.User{
		ID:                        "u1",
		Email:                     "test@example.com",
		Status:                    domain.StatusPending,
		VerificationCode:          nil,
		VerificationCodeExpiresAt: nil,
	}

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(user, nil).
		Times(1)

	_, err := uc.VerifyEmail(ctx, "test@example.com", "123456")
	if !errors.Is(err, usecase.ErrInvalidVerificationCode) {
		t.Fatalf("expected ErrInvalidVerificationCode, got: %v", err)
	}
}

func TestUserUsecase_VerifyEmail_ExpiredCode_ReturnsErrVerificationCodeExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	code := "123456"
	expired := time.Now().Add(-1 * time.Hour)
	user := &domain.User{
		ID:                        "u1",
		Email:                     "test@example.com",
		Status:                    domain.StatusPending,
		VerificationCode:          &code,
		VerificationCodeExpiresAt: &expired,
	}

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(user, nil).
		Times(1)

	_, err := uc.VerifyEmail(ctx, "test@example.com", "123456")
	if !errors.Is(err, usecase.ErrVerificationCodeExpired) {
		t.Fatalf("expected ErrVerificationCodeExpired, got: %v", err)
	}
}

func TestUserUsecase_VerifyEmail_WrongCode_ReturnsErrInvalidVerificationCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	code := "123456"
	expiresAt := time.Now().Add(15 * time.Minute)
	user := &domain.User{
		ID:                        "u1",
		Email:                     "test@example.com",
		Status:                    domain.StatusPending,
		VerificationCode:          &code,
		VerificationCodeExpiresAt: &expiresAt,
	}

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(user, nil).
		Times(1)

	_, err := uc.VerifyEmail(ctx, "test@example.com", "999999")
	if !errors.Is(err, usecase.ErrInvalidVerificationCode) {
		t.Fatalf("expected ErrInvalidVerificationCode, got: %v", err)
	}
}

func TestUserUsecase_VerifyEmail_Success_ClearsCodeAndSetsActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	uc := usecase.NewUserUsecase(repo, nil)
	ctx := context.Background()

	code := "123456"
	expiresAt := time.Now().Add(15 * time.Minute)
	user := &domain.User{
		ID:                        "u1",
		Email:                     "test@example.com",
		Status:                    domain.StatusPending,
		VerificationCode:          &code,
		VerificationCodeExpiresAt: &expiresAt,
		Profile:                   domain.Profile{Username: "Test", Phone: "123"},
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	repo.EXPECT().
		GetByEmail(ctx, "test@example.com").
		Return(user, nil).
		Times(1)

	repo.EXPECT().
		UpdateUser(ctx, gomock.AssignableToTypeOf(&domain.User{})).
		DoAndReturn(func(_ context.Context, u *domain.User) error {
			if u.VerificationCode != nil {
				t.Fatalf("expected VerificationCode to be cleared")
			}
			if u.VerificationCodeExpiresAt != nil {
				t.Fatalf("expected VerificationCodeExpiresAt to be cleared")
			}
			if u.Status != domain.StatusActive {
				t.Fatalf("expected Status active, got %q", u.Status)
			}
			if u.UpdatedAt.IsZero() {
				t.Fatalf("expected UpdatedAt to be set")
			}
			return nil
		}).
		Times(1)

	got, err := uc.VerifyEmail(ctx, "test@example.com", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Status != domain.StatusActive {
		t.Fatalf("expected user with status active, got %+v", got)
	}
}
