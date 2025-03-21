//go:generate replacer
package v1alpha

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	v1alphaUtils "github.com/weastur/maf/pkg/utils/http/api/v1alpha"
)

// Join to cluster
//
// @Summary      Join server to cluster
// @Description  Join the server to the cluster. The server becomes voter in case of success
// @Tags         raft
// @Param        request body RaftJoinRequest true "Join request"
// @Success      200 {object} Response "Response with error details or success code"
// @Router       /raft/join [post]
// @Security     ApiKeyAuth
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
func raftJoinHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	joinReq := new(RaftJoinRequest)
	if err := parseAndValidate(c, joinReq); err != nil {
		return err
	}

	if err := uCtx.co.Join(joinReq.ServerID, joinReq.Addr); err != nil {
		return err
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}

// Forget the server
//
// @Summary      Forget the server
// @Description  Forget the server. The server becomes non-voter and forgotten by the cluster in case of success
// @Description  and will not participate in the consensus
// @Tags         raft
// @Param        request body RaftForgetRequest true "Forget request"
// @Success      200 {object} Response "Response with error details or success code"
// @Router       /raft/forget [post]
// @Security     ApiKeyAuth
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
func raftForgetHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	forgetReq := new(RaftForgetRequest)
	if err := parseAndValidate(c, forgetReq); err != nil {
		return err
	}

	if err := uCtx.co.Forget(forgetReq.ServerID); err != nil {
		return err
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}

// Get raft info
//
// @Summary      Return raft info
// @Description  Return the raft cluster info with current server state and stats
// @Tags         raft
// @Success      200 {object} Response{data=RaftInfoResponse} "Raft cluster info"
// @Router       /raft/info [get]
// @Param        include_stats query bool false "Include extended stats"
// @Security     ApiKeyAuth
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
func raftInfoHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	coInfo, err := uCtx.co.GetInfo(c.QueryBool("include_stats"))
	if err != nil {
		return err
	}

	data := &RaftInfoResponse{}
	if err := copier.Copy(data, coInfo); err != nil {
		return err
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, data, nil)
}

// Get key from kv store
//
// @Summary      Return value of the key
// @Description  Return value for the key from kv store
// @Tags         raft
// @Success      200 {object} Response{data=KVGetResponse} "KV get response"
// @Router       /raft/kv/{key} [get]
// @Param        key path string true "Key to receive value"
// @Security     ApiKeyAuth
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
func raftKVGetHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	key := c.Params("key")

	value, ok := uCtx.co.Get(key)

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, &KVGetResponse{Key: key, Value: value, Exist: ok}, nil)
}

// Set key/value in kv store
//
// @Summary      Set value for key
// @Description  Set value in kv store for the key
// @Tags         raft
// @Param        request body KVSetRequest true "KV set request"
// @Success      200 {object} Response "Response with error details or success code"
// @Router       /raft/kv [post]
// @Security     ApiKeyAuth
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
func raftKVSetHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	setReq := new(KVSetRequest)
	if err := parseAndValidate(c, setReq); err != nil {
		return err
	}

	if err := uCtx.co.Set(setReq.Key, setReq.Value); err != nil {
		return err
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}

// Delete value from kv store
//
// @Summary      Delete value for key
// @Description  Delete value in kv store for the key
// @Tags         raft
// @Success      200 {object} Response "Response with error details or success code"
// @Router       /raft/kv/{key} [delete]
// @Param        key path string true "Key to delete value for"
// @Security     ApiKeyAuth
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
func raftKVDeleteHandler(c *fiber.Ctx) error {
	uCtx := unpackCtx(c)

	key := c.Params("key")

	if err := uCtx.co.Delete(key); err != nil {
		return err
	}

	return v1alphaUtils.WrapResponse(c, v1alphaUtils.StatusSuccess, nil, nil)
}
