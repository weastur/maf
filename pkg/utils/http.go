package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func RunFiberApp(app *fiber.App, addr string, certFile string, keyFile string, clientCertFile string) error {
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

func HTTPVersionHandler(c *fiber.Ctx) error {
	return c.SendString(AppVersion())
}
