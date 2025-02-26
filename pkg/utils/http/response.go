package http

import "github.com/gofiber/fiber/v2"

type ResponseStatus int

const (
	StatusSuccess ResponseStatus = iota
	StatusError
	StatusWarning
)

var statusName = map[ResponseStatus]string{
	StatusSuccess: "success",
	StatusError:   "error",
	StatusWarning: "warning",
}

func (rs ResponseStatus) String() string {
	return statusName[rs]
}

func WrapResponse(c *fiber.Ctx, status ResponseStatus, data any, err error) error {
	return c.JSON(fiber.Map{
		"status": status.String(),
		"data":   data,
		"error":  err,
	})
}
