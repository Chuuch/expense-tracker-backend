package http

type UserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	IsMFAEnabled bool   `json:"is_mfa_enabled"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken,omitempty"`
}

type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
