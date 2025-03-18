package client

import (
	"errors"
	"fmt"

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

func New(addr string) *Client {
	return &Client{
		Addr:      addr,
		AuthToken: defaultAPIKey,
	}
}

func (c *Client) Join(serverID, addr string) error {
	r := resty.New()
	defer r.Close()

	res, err := r.R().
		SetHeaderAuthorizationKey(authHeader).
		SetAuthScheme("").
		SetAuthToken(c.AuthToken).
		SetBody(&JoinRequest{
			ServerID: serverID,
			Addr:     addr,
		}).
		SetResult(&Response{}).
		Post("http://" + c.Addr + "/api/v1alpha/raft/join")
	if err != nil {
		return err
	}

	data, ok := res.Result().(*Response)
	if !ok {
		return errFailedToCastResponse
	}

	if data.Error != "" {
		return fmt.Errorf("failed to join: %s", data.Error) //nolint:err113
	}

	return nil
}
