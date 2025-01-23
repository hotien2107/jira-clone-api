package routers

import (
	"github.com/gofiber/fiber/v2"
	workspaceCtrl "jira-clone-api/api/controllers/workspace"
)

type Workspace interface {
	V1()
}
type workspace struct {
	router fiber.Router
	ctrl   workspaceCtrl.Controller
}

func NewWorkspace(router fiber.Router) Workspace {
	return &workspace{router: router.Group("/workspaces"), ctrl: workspaceCtrl.New()}
}

func (r workspace) V1() {
	r.root()
}

func (r workspace) root() {
	r.router.Post("/", r.ctrl.Create)
}
