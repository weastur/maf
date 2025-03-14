package v1alpha

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/weastur/maf/pkg/server/worker/raft"
	v1alphaUtils "github.com/weastur/maf/pkg/utils/http/api/v1alpha"
)

func joinHandler(c *fiber.Ctx) error {
	logger := zerolog.Ctx(c.UserContext())
	co, _ := c.UserContext().Value(consensusInstanceContextKey).(raft.Consensus)

	logger.Trace().Msgf("Is leader? %t", co.IsLeader())

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}

func leaveHandler(c *fiber.Ctx) error {
	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}
