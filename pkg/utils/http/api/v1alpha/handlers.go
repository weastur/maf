//go:generate replacer
package v1alpha

import (
	"github.com/gofiber/fiber/v2"
	"github.com/weastur/maf/pkg/utils"
)

// Get maf version
//
// @Summary      Return version
// @Description  Return the version of running app. Not the API version, but the application
// @Tags         aux
// @Success      200 {object} Response{data=Version} "Version"
// @Router       /version [get]
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
func VersionHandler(c *fiber.Ctx) error {
	return WrapResponse(c, StatusSuccess, &Version{Version: utils.AppVersion()}, nil)
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	return WrapResponse(c, StatusError, nil, err)
}
