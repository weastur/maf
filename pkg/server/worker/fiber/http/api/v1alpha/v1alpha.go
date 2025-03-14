package v1alpha

import (
	"context"
	"sync"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/weastur/maf/pkg/server/worker/raft"
	httpUtils "github.com/weastur/maf/pkg/utils/http"
	apiUtils "github.com/weastur/maf/pkg/utils/http/api"
	v1alphaUtils "github.com/weastur/maf/pkg/utils/http/api/v1alpha"
)

const consensusInstanceContextKey = apiUtils.UserContextKey("consensusInstance")

type Validator interface {
	Validate(data any) error
}

type APIV1Alpha struct {
	prefix    string
	version   string
	validator Validator
}

var (
	instance *APIV1Alpha
	once     sync.Once
)

func Get() *APIV1Alpha {
	once.Do(func() {
		instance = &APIV1Alpha{
			version:   "v1alpha",
			prefix:    "/v1alpha",
			validator: v1alphaUtils.NewXValidator(),
		}
	})

	return instance
}

// @title MySQL auto failover server API
// @version v1alpha
// @description Server API for MySQL auto failover. Generally must be called by CLI
// @contact.name Pavel Sapezhka
// @contact.url weastur.com
// @contact.email me@weastur.com
// @license.name Mozilla Public License Version 2.0
// @license.url https://www.mozilla.org/en-US/MPL/2.0/
// @host 127.0.0.1:7080
// @tag.name aux
// @tag.description Auxiliary endpoints
// @BasePath /api/v1alpha
// @accept json
// @produce json
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Auth-Token
// @description API key for the server. For now, only 'root' is allowed
// @externalDocs.description Find out more about MAF on GitHub
// @externalDocs.url https://github.com/weastur/maf/wiki
func (api *APIV1Alpha) Init(topRouter fiber.Router, logger zerolog.Logger, co raft.Consensus) {
	router := httpUtils.APIVersionGroup(topRouter, api.version)

	router.Use(swagger.New(swagger.Config{
		Title:    "MySQL auto failover server API, version" + api.version,
		BasePath: httpUtils.APIPrefix + api.prefix,
		FilePath: "./pkg/server/worker/fiber/http/api/v1alpha/swagger.json",
		Path:     "docs",
		CacheAge: 0,
	}))

	router.Use(func(c *fiber.Ctx) error {
		ctx := context.WithValue(context.Background(), apiUtils.APIInstanceContextKey, api)
		ctx = context.WithValue(ctx, consensusInstanceContextKey, co)
		ctx = logger.WithContext(ctx)
		c.SetUserContext(ctx)

		return c.Next()
	})
	router.Use(v1alphaUtils.AuthMiddleware())

	router.Get("/version", v1alphaUtils.VersionHandler)

	router.Post("/raft/join", joinHandler)
	router.Post("/raft/leave", leaveHandler)
}

func (api *APIV1Alpha) ErrorHandler(c *fiber.Ctx, err error) error {
	return v1alphaUtils.ErrorHandler(c, err)
}
