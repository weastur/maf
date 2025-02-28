package v1alpha

import (
	"github.com/gofiber/fiber/v2"
)

func WrapResponse(c *fiber.Ctx, status Status, data any, err error) error {
	return c.JSON(Response{
		Status: status,
		Data:   data,
		Error:  err,
	})
}
