package v1alpha

import (
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	httpUtils "github.com/weastur/maf/pkg/utils/http"
	apiUtils "github.com/weastur/maf/pkg/utils/http/api"
	v1alphaUtils "github.com/weastur/maf/pkg/utils/http/api/v1alpha"
)

type v1alpha struct {
	prefix  string
	version string
}

var apiInstance apiUtils.API

func Get() apiUtils.API {
	if apiInstance == nil {
		apiInstance = &v1alpha{
			version: "v1alpha",
			prefix:  "/v1alpha",
		}
	}

	return apiInstance
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
func (api *v1alpha) Router(topRouter fiber.Router) fiber.Router {
	router := httpUtils.APIVersionGroup(topRouter, api.version)

	router.Use(swagger.New(swagger.Config{
		Title:    "MySQL auto failover agent API, version" + api.version,
		BasePath: api.Prefix(),
		FilePath: "./pkg/agent/http/api/v1alpha/swagger.json",
		Path:     "docs",
		CacheAge: 0,
	}))

	router.Use(v1alphaUtils.AuthMiddleware())

	router.Get("/version", v1alphaUtils.VersionHandler)

	return router
}

func (api *v1alpha) Prefix() string {
	return httpUtils.APIPrefix + api.prefix
}

func (api *v1alpha) Version() string {
	return api.version
}
