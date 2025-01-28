package workspace

import (
	"path"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"jira-clone-api/api/serializers"
	"jira-clone-api/common/configure"
	"jira-clone-api/common/logging"
	"jira-clone-api/common/request"
	"jira-clone-api/common/response"
	respErr "jira-clone-api/common/response/error"
	"jira-clone-api/database/mongo/models"
	"jira-clone-api/database/mongo/queries"
	"jira-clone-api/utilities/local"
	"jira-clone-api/utilities/storage_s3"
)

var (
	cfg    = configure.GetConfig()
	logger = logging.GetLogger()
)

type Controller interface {
	Create(ctx *fiber.Ctx) error
	Search(ctx *fiber.Ctx) error
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
	image, _ := ctx.FormFile("image")
	imageName := ""
	if image != nil {
		if image.Size > 1024*1024 {
			return response.New(ctx, response.Options{
				Code: fiber.StatusBadRequest, Data: "Image size must be less than 1MB",
			})
		}
		ext := filepath.Ext(image.Filename)
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".svg" {
			return response.New(ctx, response.Options{
				Code: fiber.StatusBadRequest, Data: "Image must be png, jpg or jpeg",
			})
		}
		if err := storage_s3.GetGlobal().UploadObject(image); err != nil {
			return err
		}
		imageName = image.Filename
	}
	workspace, err := queries.NewWorkspace(ctx.Context()).Create(models.Workspace{
		Name:      requestBody.Name,
		ImageName: imageName,
		UserId:    local.New(ctx).GetUser().Id,
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

func (ctrl *controller) Search(ctx *fiber.Ctx) error {
	var (
		requestBody serializers.WorkspaceSearchBodyValidate
		totalChan   = make(chan int64, 1)
		errChan     = make(chan error, 1)
	)
	if err := ctx.BodyParser(&requestBody); err != nil {
		return response.New(ctx, response.Options{
			Code: fiber.StatusBadRequest, Data: respErr.ErrFieldWrongType,
		})
	}
	if err := requestBody.Validate(); err != nil {
		return err
	}
	userId := local.New(ctx).GetUser().Id
	go func() {
		total, err := queries.NewWorkspace(ctx.Context()).TotalByNameRegexAndUserId(requestBody.Name, userId)
		errChan <- err
		totalChan <- total
	}()
	pagination := request.NewPagination(requestBody.Limit, requestBody.Page)
	queryOption := queries.NewOptions()
	queryOption.SetPagination(pagination)
	queryOption.AddSortKey(map[string]int{"_id": -1})
	queryOption.SetOnlyFields("_id", "name", "created_at", "updated_at", "image_name")
	workspaces, err := queries.NewWorkspace(ctx.Context()).GetByNameRegexAndUserId(requestBody.Name, userId, queryOption)
	if err != nil {
		return err
	}
	if err = <-errChan; err != nil {
		return err
	}
	pagination.SetTotal(<-totalChan)
	results := make([]serializers.WorkspaceSearchResponseItem, len(workspaces))
	for i := 0; i < len(workspaces); i++ {
		results[i].Name = workspaces[i].Name
		results[i].CreatedAt = workspaces[i].CreatedAt
		results[i].UpdatedAt = workspaces[i].UpdatedAt
		results[i].ImageUrl = path.Join(cfg.S3Prefix, workspaces[i].ImageName)
		results[i].Id = workspaces[i].Id
	}
	return response.NewArrayWithPagination(ctx, results, pagination)
}
