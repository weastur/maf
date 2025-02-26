package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/weastur/maf/pkg/utils"
)

func VersionHandler(c *fiber.Ctx) error {
	return WrapResponse(c, StatusSuccess, fiber.Map{"version": utils.AppVersion()}, nil)
}
