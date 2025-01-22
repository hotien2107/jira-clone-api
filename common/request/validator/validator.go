package validator

import (
	"jira-clone-api/common/logging"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	validateEngine *validator.Validate
	logger         = logging.GetLogger()
)

type ValidateFunction struct {
	Function validator.Func
	Tag      string
}

func InitValidateEngine() *validator.Validate {
	validateEngine = validator.New()
	RegisterValidate()

	return validateEngine
}

func GetValidateEngine() *validator.Validate {
	return validateEngine
}

func RegisterValidate(functions ...ValidateFunction) {
	for _, function := range functions {
		if err := validateEngine.RegisterValidation(function.Tag, function.Function); err != nil {
			logger.Fatal().Err(err).Msg("register tag validate error")
		}
	}
}

func ParseValidateError(rawErr error) fiber.Map {
	result := fiber.Map{}
	for _, err := range rawErr.(validator.ValidationErrors) {
		msg := err.ActualTag()
		if err.Param() != "" {
			msg += "=" + err.Param()
		}
		result[err.Field()] = msg
	}
	return result
}
