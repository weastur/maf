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

	joinReq := new(JoinRequest)
	if err := c.BodyParser(joinReq); err != nil {
		logger.Error().Err(err).Msg("Failed to parse join request")

		return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusError, nil, err)
	}

	if err := co.Join(joinReq.ServerID, joinReq.Addr); err != nil {
		logger.Error().Err(err).Msg("Failed to join the consensus")

		return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusError, nil, err)
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}

func leaveHandler(c *fiber.Ctx) error {
	logger := zerolog.Ctx(c.UserContext())
	co, _ := c.UserContext().Value(consensusInstanceContextKey).(raft.Consensus)

	leaveReq := new(LeaveRequest)
	if err := c.BodyParser(leaveReq); err != nil {
		logger.Error().Err(err).Msg("Failed to parse leave request")

		return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusError, nil, err)
	}

	if err := co.Leave(leaveReq.ServerID); err != nil {
		logger.Error().Err(err).Msg("Failed to leave the consensus")

		return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusError, nil, err)
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}
