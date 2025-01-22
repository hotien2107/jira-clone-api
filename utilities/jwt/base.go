package jwt

import (
	"crypto"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"jira-clone-api/common/logging"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

var (
	global Service
	logger = logging.GetLogger()
)

type Service interface {
	InitGlobal()
	GenerateToken(tokenId string, isRefreshToken bool, duration time.Duration) (tokenStr string, err error)
	GeneratePairToken(tokenId string, accessTokenDuration time.Duration, refreshTokenDuration time.Duration) (accessToken string, refreshToken string, err error)
	ValidateToken(token string) (data *Payload, err error)
}

type service struct {
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
}

func New(privateKey, publicKey string) Service {
	return &service{
		privateKey: convertPrivateKey([]byte(privateKey)),
		publicKey:  convertPublicKey([]byte(publicKey)),
	}
}

func GetGlobal() Service {
	return global
}

func convertPrivateKey(key []byte) crypto.PrivateKey {
	signKey, err := jwt.ParseEdPrivateKeyFromPEM(key)
	if err != nil {
		logger.Fatal().Err(err).Msg("Private key format error")
	}
	return signKey
}

func convertPublicKey(key []byte) crypto.PublicKey {
	verifyKey, err := jwt.ParseEdPublicKeyFromPEM(key)
	if err != nil {
		logger.Fatal().Err(err).Msg("Public key format error")
	}
	return verifyKey
}
