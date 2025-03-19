package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/weastur/maf/pkg/utils"
	"resty.dev/v3"
)

var ErrFailedToCastResponse = errors.New("failed to cast response")

const (
	authHeader              = "X-Auth-Token"
	apiKey                  = "root" // pragma: allowlist secret
	defaultTimeout          = 10 * time.Second
	defaultRetryCount       = 3
	defaultRetryWaitTime    = 1 * time.Second
	defaultRetryMaxWaitTime = 3 * time.Second
)

type Client struct {
	Addr      string
	AuthToken string
	rclient   *resty.Client
}

func New(addr string) *Client {
	client := &Client{
		Addr:      addr,
		AuthToken: apiKey,
		rclient:   resty.New(),
	}
	client.rclient.SetHeaderAuthorizationKey(authHeader)
	client.rclient.SetAuthScheme("")
	client.rclient.SetAuthToken(client.AuthToken)
	client.rclient.SetHeader("User-Agent", "maf/"+utils.AppVersion())
	client.rclient.SetTimeout(defaultTimeout)
	client.rclient.SetRetryCount(defaultRetryCount)
	client.rclient.SetRetryWaitTime(defaultRetryWaitTime)
	client.rclient.SetRetryMaxWaitTime(defaultRetryMaxWaitTime)

	return client
}

func (c *Client) Close() {
	c.rclient.Close()
}

func (c *Client) Join(serverID, addr string) error {
	res, err := c.rclient.R().
		SetBody(&joinRequest{
			ServerID: serverID,
			Addr:     addr,
		}).
		SetResult(&response{}).
		Post("http://" + c.Addr + "/api/v1alpha/raft/join")
	if err != nil {
		return err
	}

	data, ok := res.Result().(*response)
	if !ok {
		return ErrFailedToCastResponse
	}

	if data.Error != "" {
		return fmt.Errorf("failed to join: %s", data.Error) //nolint:err113
	}

	return nil
}
