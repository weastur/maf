package v1alpha

import (
	"context"
	"os"
	"sync"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/weastur/maf/internal/server/worker/raft"
	httpUtils "github.com/weastur/maf/internal/utils/http"
	apiUtils "github.com/weastur/maf/internal/utils/http/api"
	v1alphaUtils "github.com/weastur/maf/internal/utils/http/api/v1alpha"
)

const (
	consensusInstanceContextKey = apiUtils.UserContextKey("consensusInstance")
	swaggerFilePath             = "./internal/server/worker/fiber/http/api/v1alpha/swagger.json"
)

type Consensus interface {
	IsLeader() bool
	Join(serverID, addr string) error
	Forget(serverID string) error
	GetInfo(verbose bool) (*raft.Info, error)
	Get(key string) (string, bool)
	Set(key, value string) error
	Delete(key string) error
}

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
// @tag.name raft
// @tag.description Raft-related endpoints
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
func (api *APIV1Alpha) Init(topRouter fiber.Router, logger zerolog.Logger, co Consensus) {
	router := httpUtils.APIVersionGroup(topRouter, api.version)
	if _, err := os.Stat(swaggerFilePath); !os.IsNotExist(err) {
		router.Use(swagger.New(swagger.Config{
			Title:    "MySQL auto failover server API, version" + api.version,
			BasePath: httpUtils.APIPrefix + api.prefix,
			FilePath: swaggerFilePath,
			Path:     "docs",
			CacheAge: 0,
		}))
	}

	router.Use(func(c *fiber.Ctx) error {
		ctx := context.WithValue(context.Background(), apiUtils.APIInstanceContextKey, api)
		ctx = context.WithValue(ctx, consensusInstanceContextKey, co)
		ctx = logger.WithContext(ctx)
		c.SetUserContext(ctx)

		return c.Next()
	})
	router.Use(v1alphaUtils.AuthMiddleware())

	router.Get("/version", v1alphaUtils.VersionHandler)

	router.Post("/raft/join", raftJoinHandler)
	router.Post("/raft/forget", raftForgetHandler)
	router.Get("/raft/info", raftInfoHandler)
	router.Get("/raft/kv/:key", raftKVGetHandler)
	router.Post("/raft/kv", raftKVSetHandler)
	router.Delete("/raft/kv/:key", raftKVDeleteHandler)
}

func (api *APIV1Alpha) ErrorHandler(c *fiber.Ctx, err error) error {
	return v1alphaUtils.ErrorHandler(c, err)
}
