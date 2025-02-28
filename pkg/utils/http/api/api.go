package api

import "github.com/gofiber/fiber/v2"

type API interface {
	Router(topRouter fiber.Router) fiber.Router
	Prefix() string
	Version() string
}
