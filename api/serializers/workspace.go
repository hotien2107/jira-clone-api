package serializers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"jira-clone-api/common/request/validator"
	"jira-clone-api/common/response"
)

type WorkspaceCreateBodyValidate struct {
	Name string `form:"name" validate:"required"`
}

func (v *WorkspaceCreateBodyValidate) Validate() error {
	validateEngine := validator.GetValidateEngine()
	if err := validateEngine.Struct(v); err != nil {
		return response.NewError(fiber.StatusBadRequest, response.ErrorOptions{
			Data: validator.ParseValidateError(err),
		})
	}
	return nil
}

type WorkspaceSearchBodyValidate struct {
	Name  string `json:"name" validate:"omitempty"`
	Page  int64  `json:"page" validate:"omitempty"`
	Limit int64  `json:"limit" validate:"omitempty"`
}

func (v *WorkspaceSearchBodyValidate) Validate() error {
	validateEngine := validator.GetValidateEngine()
	if err := validateEngine.Struct(v); err != nil {
		return response.NewError(fiber.StatusBadRequest, response.ErrorOptions{
			Data: validator.ParseValidateError(err),
		})
	}
	return nil
}

type WorkspaceSearchResponseItem struct {
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	ImageUrl  string             `json:"image_url"`
	Name      string             `json:"name"`
	Id        primitive.ObjectID `json:"id"`
}
