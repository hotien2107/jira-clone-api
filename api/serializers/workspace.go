package serializers

import (
	"github.com/gofiber/fiber/v2"
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
