package client

import (
	"errors"
	"fmt"

	"github.com/weastur/maf/pkg/server/worker/fiber/http/api/v1alpha"
	v1alphaUtils "github.com/weastur/maf/pkg/utils/http/api/v1alpha"
	"resty.dev/v3"
)

var errFailedToCastResponse = errors.New("failed to cast response")

const (
	authHeader    = "X-Auth-Token"
	defaultAPIKey = "root" // pragma: allowlist secret
)

type Client struct {
	Addr      string
	AuthToken string
}

func NewClient(addr string) *Client {
	return &Client{
		Addr:      addr,
		AuthToken: defaultAPIKey,
	}
}

func (c *Client) JoinReqiuest(serverID, addr string) error {
	r := resty.New()
	defer r.Close()

	res, err := r.R().
		SetHeaderAuthorizationKey(authHeader).
		SetAuthToken(c.AuthToken).
		SetBody(&v1alpha.JoinRequest{
			ServerID: serverID,
			Addr:     addr,
		}).
		SetResult(&v1alphaUtils.Response{}).
		Post(c.Addr + "/api/v1alpha/server/join")
	if err != nil {
		return err
	}

	data, ok := res.Result().(*v1alphaUtils.Response)
	if !ok {
		return errFailedToCastResponse
	}

	if data.Error != nil {
		return fmt.Errorf("error: %w", data.Error)
	}

	return nil
}
