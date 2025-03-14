package v1alpha

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/weastur/maf/pkg/server/worker/raft"
	apiUtils "github.com/weastur/maf/pkg/utils/http/api"
)

func unpackCtx(c *fiber.Ctx) (*zerolog.Logger, raft.Consensus, *APIV1Alpha) {
	logger := zerolog.Ctx(c.UserContext())
	co, _ := c.UserContext().Value(consensusInstanceContextKey).(raft.Consensus)
	api, _ := c.UserContext().Value(apiUtils.APIInstanceContextKey).(*APIV1Alpha)

	return logger, co, api
}
