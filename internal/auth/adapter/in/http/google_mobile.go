package http

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type googleIDTokenInfo struct {
	Email     string
	Name      string
	FirstName string
	LastName  string
}

func verifyGoogleIDToken(ctx context.Context, rawToken, audience string) (*googleIDTokenInfo, error) {
	payload, err := idtoken.Validate(ctx, rawToken, audience)
	if err != nil {
		return nil, fmt.Errorf("google id token validate: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	firstName, _ := payload.Claims["given_name"].(string)
	lastName, _ := payload.Claims["family_name"].(string)

	if email == "" {
		return nil, fmt.Errorf("google id token missing email")
	}

	return &googleIDTokenInfo{
		Email:     email,
		Name:      name,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}
