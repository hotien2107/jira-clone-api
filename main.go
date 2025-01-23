package main

import (
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "go.uber.org/automaxprocs"
	"jira-clone-api/api/routers"
	"jira-clone-api/common/configure"
	"jira-clone-api/common/logging"
	"jira-clone-api/common/request/validator"
	"jira-clone-api/common/response"
	respErr "jira-clone-api/common/response/error"
	"jira-clone-api/database"
	"jira-clone-api/utilities/jwt"
	"jira-clone-api/utilities/storage_s3"
)

var cfg = configure.GetConfig()

func main() {
	logging.InitLogger()
	validator.InitValidateEngine()
	database.InitDatabase()
	jwt.New(cfg.TokenPrivateKey, cfg.TokenPublicKey).InitGlobal()
	storage_s3.New().InitGlobal()
	app := fiber.New(fiber.Config{
		ErrorHandler: response.FiberErrorHandler,
		JSONDecoder:  sonic.Unmarshal,
		JSONEncoder:  sonic.Marshal,
		BodyLimit:    cfg.APIBodyLimitSize,
	})
	addMiddleware(app)
	addV1Route(app)
	handleURLNotFound(app)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		logging.GetLogger().Info().Msg("ready")
		if err := app.Listen(cfg.ServerAddress()); err != nil {
			logging.GetLogger().Error().Err(err).Str("function", "main").Str("functionInline", "app.Listen").Msg("Can't start server")
		}
		sigChan <- syscall.SIGTERM
	}()
	<-sigChan
	logging.GetLogger().Info().Msg("Shutting down...")
	_ = app.Shutdown()
	database.DisconnectDatabase()
}

func handleURLNotFound(app *fiber.App) {
	app.Use(func(ctx *fiber.Ctx) error {
		return response.New(ctx, response.Options{Code: fiber.StatusNotFound, Data: respErr.ErrUrlNotFound})
	})
}

func addMiddleware(app *fiber.App) {
	app.Use(cors.New())
	if cfg.ElasticAPMEnable {
		app.Use(logging.FiberApmMiddleware())
	} else {
		recoverConfig := recover.ConfigDefault
		recoverConfig.EnableStackTrace = cfg.Debug
		app.Use(recover.New(recoverConfig))
	}
	app.Use(logging.FiberLoggerMiddleware())
}

func addV1Route(app *fiber.App) {
	route := app.Group("/api/jira-clone-api/v1")
	routers.NewAuthenticate(route).V1()
	routers.NewWorkspace(route).V1()
}
