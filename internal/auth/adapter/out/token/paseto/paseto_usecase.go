package paseto

import (
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/chuuch/expense-tracker-backend/internal/auth/domain"
)

type PasetoUsecase struct {
	key paseto.V4SymmetricKey
}

func NewPasetoUsecase(symmetricKey string) (*PasetoUsecase, error) {
	if len(symmetricKey) != 32 {
		return nil, fmt.Errorf("invalid paseto key length")
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(symmetricKey))
	if err != nil {
		return nil, fmt.Errorf("invalid paseto key: %w", err)
	}

	return &PasetoUsecase{key: key}, nil
}

func (u *PasetoUsecase) GenerateToken(user *domain.User, duration time.Duration) (string, error) {
	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(duration))
	token.SetString("user_id", user.ID)
	token.SetString("role", string(user.Role))

	return token.V4Encrypt(u.key, nil), nil
}

func (u *PasetoUsecase) VerifyToken(signedToken string) (string, error) {
	parser := paseto.NewParser()

	token, err := parser.ParseV4Local(u.key, signedToken, nil)
	if err != nil {
		return "", fmt.Errorf("token verification failed: %w", err)
	}

	userID, err := token.GetString("user_id")
	if err != nil {
		return "", fmt.Errorf("user_id claim not found")
	}

	return userID, nil

}
