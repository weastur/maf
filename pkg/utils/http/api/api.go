package api

import "github.com/gofiber/fiber/v2"

type UserContextKey string

type API interface {
	Router(topRouter fiber.Router) fiber.Router
	Prefix() string
	Version() string
	ErrorHandler(c *fiber.Ctx, err error) error
}
