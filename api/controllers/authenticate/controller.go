package authenticate

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"jira-clone-api/api/serializers"
	"jira-clone-api/common/configure"
	"jira-clone-api/common/logging"
	"jira-clone-api/common/response"
	respErr "jira-clone-api/common/response/error"
	"jira-clone-api/database/mongo/models"
	"jira-clone-api/database/mongo/queries"
	"jira-clone-api/utilities/jwt"
	"jira-clone-api/utilities/local"
)

var (
	cfg    = configure.GetConfig()
	logger = logging.GetLogger()
)

type Controller interface {
	Login(ctx *fiber.Ctx) error
	Register(ctx *fiber.Ctx) error
	GetUserInfo(ctx *fiber.Ctx) error
}

type controller struct {
	service serviceInterface
}

func New() Controller {
	return &controller{
		service: newService(),
	}
}

func (ctrl *controller) Register(ctx *fiber.Ctx) error {
	var requestBody serializers.AuthenticateRegisterBodyValidate
	if err := ctx.BodyParser(&requestBody); err != nil {
		return response.New(ctx, response.Options{
			Code: fiber.StatusBadRequest, Data: respErr.ErrFieldWrongType,
		})
	}
	if err := requestBody.Validate(); err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Str("function", "Login").Str("functionInline", "jwt.GetGlobal().GenerateFromPassword").Msg("authenticateController")
		return response.New(ctx, response.Options{Code: fiber.StatusInternalServerError})

	}
	user, err := queries.NewUser(ctx.Context()).Create(models.User{
		Username: requestBody.Username,
		Password: string(hashedPassword),
		Email:    requestBody.Email,
	})
	if err != nil {
		return err
	}
	return response.New(ctx, response.Options{
		Code: fiber.StatusOK,
		Data: fiber.Map{
			"id": user.Id,
		},
	})
}

func (ctrl *controller) Login(ctx *fiber.Ctx) error {
	var requestBody serializers.AuthenticateLoginBodyValidate
	if err := ctx.BodyParser(&requestBody); err != nil {
		return response.New(ctx, response.Options{
			Code: fiber.StatusBadRequest, Data: respErr.ErrFieldWrongType,
		})
	}
	if err := requestBody.Validate(); err != nil {
		return err
	}
	optionQuery := queries.NewOptions()
	optionQuery.SetOnlyFields("password")
	user, err := queries.NewUser(ctx.Context()).GetByUsername(requestBody.Username, optionQuery)
	if err != nil {
		return err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(requestBody.Password)); err != nil {
		logger.Error().Err(err).Str("function", "Login").Str("functionInline", "jwt.GetGlobal().CompareHashAndPassword").Msg("authenticateController")
		return response.New(ctx, response.Options{Code: fiber.StatusUnauthorized, Data: "Invalid password"})
	}
	tokenId, err := queries.NewToken(ctx.Context()).Create(models.Token{
		ExpiredAt: time.Now().Add(time.Hour * 5),
		UserId:    user.Id,
	})
	if err != nil {
		return err
	}
	accessToken, _, err := jwt.GetGlobal().GeneratePairToken(tokenId.Hex(), cfg.AccessTokenTimeout, cfg.RefreshTokenTimeout)
	if err != nil {
		logger.Error().Err(err).Str("function", "Login").Str("functionInline", "jwt.GetGlobal().GeneratePairToken").Msg("authenticateController")
		return response.New(ctx, response.Options{Code: fiber.StatusInternalServerError})
	}
	return response.New(ctx, response.Options{
		Code: fiber.StatusOK,
		Data: serializers.AuthenticateLoginResponse{
			AccessToken: accessToken,
			TokenType:   cfg.TokenType,
		},
	})
}

func (ctrl *controller) GetUserInfo(ctx *fiber.Ctx) error {
	user := local.New(ctx).GetUser()
	return response.New(ctx, response.Options{
		Code: fiber.StatusOK,
		Data: serializers.AuthenticateGetUserInfoResponse{
			Username: user.Username,
			Email:    user.Email,
			Id:       user.Id,
		},
	})
}
