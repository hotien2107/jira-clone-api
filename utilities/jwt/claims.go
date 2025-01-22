package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type Payload struct {
	jwt.RegisteredClaims
	TokenType string `json:"token_type"`
}

func (p Payload) Valid() error {
	return jwt.NewValidator().Validate(p.RegisteredClaims)
}

func (p Payload) IsAccessToken() bool {
	return p.TokenType == TokenTypeAccess
}

func (p Payload) IsRefreshToken() bool {
	return p.TokenType == TokenTypeRefresh
}

type payloadTransaction struct {
	jwt.RegisteredClaims
}
