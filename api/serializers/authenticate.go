package serializers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"jira-clone-api/common/request/validator"
	"jira-clone-api/common/response"
)

type AuthenticateRegisterBodyValidate struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (v *AuthenticateRegisterBodyValidate) Validate() error {
	validateEngine := validator.GetValidateEngine()
	if err := validateEngine.Struct(v); err != nil {
		return response.NewError(fiber.StatusBadRequest, response.ErrorOptions{
			Data: validator.ParseValidateError(err),
		})
	}
	return nil
}

type AuthenticateLoginBodyValidate struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (v *AuthenticateLoginBodyValidate) Validate() error {
	validateEngine := validator.GetValidateEngine()
	if err := validateEngine.Struct(v); err != nil {
		return response.NewError(fiber.StatusBadRequest, response.ErrorOptions{
			Data: validator.ParseValidateError(err),
		})
	}
	return nil
}

type AuthenticateLoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type AuthenticateGetUserInfoResponse struct {
	Username string             `json:"username"`
	Email    string             `json:"email"`
	Id       primitive.ObjectID `json:"id"`
}
