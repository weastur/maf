package v1alpha

import (
	"fmt"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/weastur/maf/pkg/server/worker/raft"
	apiUtils "github.com/weastur/maf/pkg/utils/http/api"
)

const requestIDLogField = fiberzerolog.FieldRequestID

type unwrappedCtx struct {
	logger zerolog.Logger
	co     raft.Consensus
	api    *APIV1Alpha
	rid    string
}

func unpackCtx(c *fiber.Ctx) *unwrappedCtx {
	logger := zerolog.Ctx(c.UserContext())
	co, _ := c.UserContext().Value(consensusInstanceContextKey).(raft.Consensus)
	api, _ := c.UserContext().Value(apiUtils.APIInstanceContextKey).(*APIV1Alpha)
	rid, _ := c.UserContext().Value(apiUtils.RequestIDContextKey).(string)

	return &unwrappedCtx{
		logger: logger.With().Str(requestIDLogField, rid).Logger(),
		co:     co,
		api:    api,
		rid:    rid,
	}
}

func parseAndValidate(c *fiber.Ctx, req any) error {
	uCtx := unpackCtx(c)

	if err := c.BodyParser(req); err != nil {
		uCtx.logger.Error().Err(err).Msg("Failed to parse request")

		return fmt.Errorf("failed to parse request: %w", err)
	}

	if err := uCtx.api.validator.Validate(req); err != nil {
		uCtx.logger.Error().Err(err).Msg("Failed to validate request")

		return fmt.Errorf("failed to validate request: %w", err)
	}

	return nil
}
