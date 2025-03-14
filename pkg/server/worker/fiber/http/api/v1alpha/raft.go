package v1alpha

import (
	"github.com/gofiber/fiber/v2"
	v1alphaUtils "github.com/weastur/maf/pkg/utils/http/api/v1alpha"
)

func joinHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	joinReq := new(JoinRequest)
	if err := parseAndValidate(c, joinReq); err != nil {
		return err
	}

	if err := uCtx.co.Join(joinReq.ServerID, joinReq.Addr); err != nil {
		return err
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}

func leaveHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	leaveReq := new(LeaveRequest)
	if err := parseAndValidate(c, leaveReq); err != nil {
		return err
	}

	if err := uCtx.co.Leave(leaveReq.ServerID); err != nil {
		return err
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}
