package v1alpha

import (
	"context"
	"sync"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	httpUtils "github.com/weastur/maf/internal/utils/http"
	apiUtils "github.com/weastur/maf/internal/utils/http/api"
	v1alphaUtils "github.com/weastur/maf/internal/utils/http/api/v1alpha"
)

type APIV1Alpha struct {
	prefix  string
	version string
}

var (
	instance *APIV1Alpha
	once     sync.Once
)

func Get() *APIV1Alpha {
	once.Do(func() {
		instance = &APIV1Alpha{
			version: "v1alpha",
			prefix:  "/v1alpha",
		}
	})

	return instance
}

// @title MySQL auto failover agent API
// @version v1alpha
// @description Agent API for MySQL auto failover. Generally must be called by CLI or server
// @contact.name Pavel Sapezhka
// @contact.url weastur.com
// @contact.email me@weastur.com
// @license.name Mozilla Public License Version 2.0
// @license.url https://www.mozilla.org/en-US/MPL/2.0/
// @host 127.0.0.1:7070
// @tag.name aux
// @tag.description Auxiliary endpoints
// @BasePath /api/v1alpha
// @accept json
// @produce json
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Auth-Token
// @description API key for the agent. For now, only 'root' is allowed
// @externalDocs.description Find out more about MAF on GitHub
// @externalDocs.url https://github.com/weastur/maf/wiki
func (api *APIV1Alpha) Init(topRouter fiber.Router, logger zerolog.Logger) {
	router := httpUtils.APIVersionGroup(topRouter, api.version)

	router.Use(swagger.New(swagger.Config{
		Title:    "MySQL auto failover agent API, version" + api.version,
		BasePath: httpUtils.APIPrefix + api.prefix,
		FilePath: "./internal/agent/worker/fiber/http/api/v1alpha/swagger.json",
		Path:     "docs",
		CacheAge: 0,
	}))

	router.Use(func(c *fiber.Ctx) error {
		ctx := context.WithValue(context.Background(), apiUtils.APIInstanceContextKey, api)
		ctx = logger.WithContext(ctx)
		c.SetUserContext(ctx)

		return c.Next()
	})

	router.Use(v1alphaUtils.AuthMiddleware())

	router.Get("/version", v1alphaUtils.VersionHandler)
}

func (api *APIV1Alpha) ErrorHandler(c *fiber.Ctx, err error) error {
	return v1alphaUtils.ErrorHandler(c, err)
}
