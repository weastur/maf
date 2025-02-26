package http

import "github.com/gofiber/fiber/v2"

type Healthchecker interface {
	IsLive(c *fiber.Ctx) bool
	IsReady(c *fiber.Ctx) bool
}
