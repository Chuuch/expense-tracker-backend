package domain

type TokenClaims struct {
	UserID string
	Role   UserRole
}
