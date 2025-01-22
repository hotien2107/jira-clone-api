package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (s *service) InitGlobal() {
	global = s
}

func (s *service) GenerateToken(tokenId string, isRefreshToken bool, duration time.Duration) (tokenStr string, err error) {
	if tokenId == "" {
		return "", errors.New("tokenId is empty")
	}
	tokenType := TokenTypeAccess
	if isRefreshToken {
		tokenType = TokenTypeRefresh
	}
	payload := Payload{
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			ID:        tokenId,
		},
		TokenType: tokenType,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, payload)
	return token.SignedString(s.privateKey)
}

func (s *service) GeneratePairToken(tokenId string, accessTokenDuration time.Duration, refreshTokenDuration time.Duration) (accessToken string, refreshToken string, err error) {
	accessToken, err = s.GenerateToken(tokenId, false, accessTokenDuration)
	if err != nil {
		return accessToken, refreshToken, err
	}
	refreshToken, err = s.GenerateToken(tokenId, true, refreshTokenDuration)
	return accessToken, refreshToken, err
}

func (s *service) ValidateToken(token string) (data *Payload, err error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodEd25519)
		if !ok {
			return nil, errors.New("token validate failed")
		}
		return s.publicKey, nil
	}

	payload := new(Payload)
	if _, err = jwt.ParseWithClaims(token, payload, keyFunc); err != nil {
		return nil, err
	}
	return payload, nil
}
