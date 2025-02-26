package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

const (
	defaultRateLimit           = 100
	defaultRateLimitExpiration = 30 * time.Second
)

func Listen(app *fiber.App, addr string, certFile string, keyFile string, clientCertFile string) error {
	switch {
	case clientCertFile != "":
		fmt.Printf("Listening with mutual TLS on %s\n", addr)

		if err := app.ListenMutualTLS(addr, certFile, keyFile, clientCertFile); err != nil {
			return fmt.Errorf("failed to listen with mutual TLS: %w", err)
		}
	case certFile != "" && keyFile != "":
		fmt.Printf("Listening with TLS on %s\n", addr)

		if err := app.ListenTLS(addr, certFile, keyFile); err != nil {
			return fmt.Errorf("failed to listen with TLS: %w", err)
		}
	default:
		fmt.Printf("Listening on %s\n", addr)

		if err := app.Listen(addr); err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
	}

	return nil
}

func APIGroup(app *fiber.App) fiber.Router {
	return app.Group("/api", func(c *fiber.Ctx) error {
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
