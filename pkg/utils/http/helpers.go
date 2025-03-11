package http

import (
	"fmt"
	"time"

	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/weastur/maf/pkg/utils"
	apiUtils "github.com/weastur/maf/pkg/utils/http/api"
)

const (
	defaultRateLimit           = 100
	defaultRateLimitExpiration = 30 * time.Second
	APIPrefix                  = "/api"
)

type API interface {
	ErrorHandler(c *fiber.Ctx, err error) error
}

func Listen(app *fiber.App, addr string, certFile string, keyFile string, clientCertFile string) error {
	switch {
	case clientCertFile != "":
		log.Info().Msgf("Listening with mutual TLS on %s", addr)

		if err := app.ListenMutualTLS(addr, certFile, keyFile, clientCertFile); err != nil {
			return fmt.Errorf("failed to listen with mutual TLS: %w", err)
		}
	case certFile != "" && keyFile != "":
		log.Info().Msgf("Listening with TLS on %s", addr)

		if err := app.ListenTLS(addr, certFile, keyFile); err != nil {
			return fmt.Errorf("failed to listen with TLS: %w", err)
		}
	default:
		log.Info().Msgf("Listening on %s", addr)

		if err := app.Listen(addr); err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
	}

	return nil
}

func APIGroup(app *fiber.App) fiber.Router {
	return app.Group(APIPrefix, func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		return c.Next()
	})
}

func APIVersionGroup(api fiber.Router, version string) fiber.Router {
	return api.Group("/"+version, func(c *fiber.Ctx) error {
		c.Set("X-API-Version", version)

		return c.Next()
	})
}

func AttachGenericMiddlewares(app *fiber.App, healthchecker Healthchecker) {
	app.Use(fibersentry.New(fibersentry.Config{
		Repanic:         true,
		WaitForDelivery: true,
	}))
	app.Use(func(c *fiber.Ctx) error {
		if hub := fibersentry.GetHubFromContext(c); hub != nil {
			hub.Scope().SetTag(utils.SentryScopeTag, "fiber")
		}

		return c.Next()
	})

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &log.Logger,
		Fields: []string{"requestId", "ip", "method", "path", "status", "latency"},
		Levels: []zerolog.Level{zerolog.ErrorLevel, zerolog.ErrorLevel, zerolog.DebugLevel},
	}))
	app.Use(compress.New())
	app.Use(requestid.New())
	app.Use(limiter.New(limiter.Config{
		Max:        defaultRateLimit,
		Expiration: defaultRateLimitExpiration,
	}))
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe:  healthchecker.IsLive,
		ReadinessProbe: healthchecker.IsReady,
	}))
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	apiInstance, ok := c.UserContext().Value(apiUtils.UserContextKey("apiInstance")).(API)
	if ok {
		return apiInstance.ErrorHandler(c, err)
	}

	return fiber.DefaultErrorHandler(c, err)
}
