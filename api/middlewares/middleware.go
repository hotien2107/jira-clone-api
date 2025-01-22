package authenticate

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"jira-clone-api/common/configure"
	"jira-clone-api/common/response"
	respErr "jira-clone-api/common/response/error"
	"jira-clone-api/database/mongo/queries"
	jwtTool "jira-clone-api/utilities/jwt"
	"jira-clone-api/utilities/local"
)

var cfg = configure.GetConfig()

func RefreshToken(ctx *fiber.Ctx) error {
	tokenString := ctx.Get("Authorization")
	if tokenString == "" {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenRequired})
	}
	if !strings.HasPrefix(tokenString, cfg.TokenType) {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrongFormat})
	}
	tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, cfg.TokenType))
	payload, err := jwtTool.GetGlobal().ValidateToken(tokenString)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrong})
	}
	if !payload.IsRefreshToken() {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrong})
	}
	tokenId, err := primitive.ObjectIDFromHex(payload.ID)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrong})
	}
	queryOption := queries.NewOptions()
	queryOption.SetOnlyFields("_id", "user_id")
	token, err := queries.NewToken(ctx.Context()).GetById(tokenId, queryOption)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenRevoked})
	}
	queryOption.SetOnlyFields("_id", "email", "username")
	user, err := queries.NewUser(ctx.Context()).GetById(token.UserId, queryOption)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: "User not found"})
	}
	localService := local.New(ctx)
	localService.SetUser(*user)
	localService.SetTokenId(tokenId)
	return ctx.Next()
}

func AccessToken(ctx *fiber.Ctx) error {
	tokenString := ctx.Get("Authorization")
	if tokenString == "" {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenRequired})
	}
	if !strings.HasPrefix(tokenString, cfg.TokenType) {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrongFormat})
	}
	token := strings.TrimSpace(strings.TrimPrefix(tokenString, cfg.TokenType))
	payload, err := jwtTool.GetGlobal().ValidateToken(token)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrong})
	}
	if !payload.IsAccessToken() {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrong})
	}
	tokenId, err := primitive.ObjectIDFromHex(payload.ID)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenWrong})
	}
	opt := queries.NewOptions()
	opt.SetOnlyFields("_id", "user_id")
	tok, err := queries.NewToken(ctx.Context()).GetById(tokenId, opt)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: respErr.ErrTokenRevoked})
	}
	opt.SetOnlyFields("_id", "email", "username")
	user, err := queries.NewUser(ctx.Context()).GetById(tok.UserId, opt)
	if err != nil {
		return response.NewError(fiber.StatusUnauthorized, response.ErrorOptions{Data: "User not found"})
	}
	local.New(ctx).SetUser(*user)
	return ctx.Next()
}
