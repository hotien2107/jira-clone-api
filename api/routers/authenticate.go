package routers

import (
	"github.com/gofiber/fiber/v2"
	authenticateCtrl "jira-clone-api/api/controllers/authenticate"
	authMiddleware "jira-clone-api/api/middlewares"
)

type Authenticate interface {
	V1()
}
type authenticate struct {
	router fiber.Router
	ctrl   authenticateCtrl.Controller
}

func NewAuthenticate(router fiber.Router) Authenticate {
	return &authenticate{router: router.Group("/auth"), ctrl: authenticateCtrl.New()}
}

func (r authenticate) V1() {
	r.root()
}

func (r authenticate) root() {
	r.router.Post("/register", r.ctrl.Register)
	r.router.Post("/login", r.ctrl.Login)
	r.router.Get("/user-info", authMiddleware.AccessToken, r.ctrl.GetUserInfo)
}
