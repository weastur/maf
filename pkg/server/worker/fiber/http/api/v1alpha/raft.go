package v1alpha

import (
	"github.com/gofiber/fiber/v2"
	v1alphaUtils "github.com/weastur/maf/pkg/utils/http/api/v1alpha"
)

// Join to cluster
//
// @Summary      Join server to cluster
// @Description  Join the server to the cluster. The server becomes voter in case of success
// @Tags         raft
// @Param        request body JoinRequest true "Join request"
// @Success      200 {object} Response "Response with error details or success code"
// @Router       /raft/join [post]
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
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

// Leave the cluster
//
// @Summary      Leave the cluster
// @Description  Leave the cluster. The server becomes non-voter in case of success
// @Description  and will not participate in the consensus
// @Tags         raft
// @Param        request body LeaveRequest true "Leave request"
// @Success      200 {object} Response "Response with error details or success code"
// @Router       /raft/leave [post]
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
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
