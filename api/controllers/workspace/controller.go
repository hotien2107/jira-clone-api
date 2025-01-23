package workspace

import (
	"github.com/gofiber/fiber/v2"
	"jira-clone-api/api/serializers"
	"jira-clone-api/common/configure"
	"jira-clone-api/common/logging"
	"jira-clone-api/common/response"
	respErr "jira-clone-api/common/response/error"
	"jira-clone-api/database/mongo/models"
	"jira-clone-api/database/mongo/queries"
	"jira-clone-api/utilities/local"
)

var (
	cfg    = configure.GetConfig()
	logger = logging.GetLogger()
)

type Controller interface {
	Create(ctx *fiber.Ctx) error
}

type controller struct {
	service serviceInterface
}

func New() Controller {
	return &controller{
		service: newService(),
	}
}

func (ctrl *controller) Create(ctx *fiber.Ctx) error {
	var requestBody serializers.WorkspaceCreateBodyValidate
	if err := ctx.BodyParser(&requestBody); err != nil {
		return response.New(ctx, response.Options{
			Code: fiber.StatusBadRequest, Data: respErr.ErrFieldWrongType,
		})
	}
	workspace, err := queries.NewWorkspace(ctx.Context()).Create(models.Workspace{
		Name:   requestBody.Name,
		UserId: local.New(ctx).GetUser().Id,
	})
	if err != nil {
		return err
	}
	return response.New(ctx, response.Options{
		Code: fiber.StatusOK,
		Data: fiber.Map{
			"id": workspace.Id,
		},
	})
}
