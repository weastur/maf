package client

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/pkg/utils"
	"github.com/weastur/maf/pkg/utils/logging"
	"resty.dev/v3"
)

const (
	authHeader                   = "X-Auth-Token"
	apiKey                       = "root" // pragma: allowlist secret
	defaultTimeout               = 10 * time.Second
	defaultRetryCount            = 3
	defaultRetryWaitTime         = 1 * time.Second
	defaultRetryMaxWaitTime      = 3 * time.Second
	defaultCircuitBreakerTimeout = 10 * time.Second
)

type Client struct {
	Host      string
	urlPerfix string
	AuthToken string
	rclient   *resty.Client
	logger    zerolog.Logger
}

func New(host string) *Client {
	client := &Client{
		Host:      host,
		AuthToken: apiKey,
		rclient:   resty.New(),
		urlPerfix: host + "/api/v1alpha",
		logger:    log.With().Str(logging.ComponentCtxKey, "server-client").Logger(),
	}
	cb := resty.NewCircuitBreaker().SetTimeout(defaultCircuitBreakerTimeout)

	client.rclient.SetHeaderAuthorizationKey(authHeader)
	client.rclient.SetAuthScheme("")
	client.rclient.SetAuthToken(client.AuthToken)
	client.rclient.SetHeader("User-Agent", "maf/"+utils.AppVersion())
	client.rclient.SetTimeout(defaultTimeout)
	client.rclient.SetRetryCount(defaultRetryCount)
	client.rclient.SetRetryWaitTime(defaultRetryWaitTime)
	client.rclient.SetRetryMaxWaitTime(defaultRetryMaxWaitTime)
	client.rclient.SetCircuitBreaker(cb)
	client.rclient.SetLogger(NewRestyLogger(client.logger))
	client.rclient.AddContentDecompresser("br", decompressBrotli)

	if e := client.logger.Debug(); e.Enabled() {
		e.Msg("Request debug enabled")
		client.rclient.EnableDebug()
	}

	if e := client.logger.Trace(); e.Enabled() {
		e.Msg("Request tracing enabled")
		client.rclient.EnableTrace()
	}

	return client
}

func NewWithTLS(host, serverCertFile string) *Client {
	client := New(host)

	client.rclient.SetRootCertificates(serverCertFile)

	return client
}

func NewWithMutualTLS(host, certFile, keyFile, serverCertFile string) *Client {
	client := New(host)

	client.rclient.SetRootCertificates(serverCertFile)
	client.rclient.SetCertificateFromFile(certFile, keyFile)

	return client
}

func NewWithAutoTLS(host string, config *TLSConfig) *Client {
	if config == nil {
		return New(host)
	} else if config.CertFile == "" && config.KeyFile == "" {
		return NewWithTLS(host, config.ServerCertFile)
	}

	return NewWithMutualTLS(host, config.CertFile, config.KeyFile, config.ServerCertFile)
}

func (c *Client) Close() {
	c.rclient.Close()
}

func (c *Client) parseResponse(res *resty.Response) (any, error) {
	if res.IsError() {
		err := &StatusCodeError{Code: res.StatusCode()}

		return nil, err
	}

	data, ok := res.Result().(*response)
	if !ok {
		return nil, ErrUnknownResponseFormat
	}

	if !data.IsSuccess() {
		err := &ServerError{Details: data.Error}

		return nil, err
	}

	return data.Data, nil
}

func (c *Client) Join(serverID, addr string) error {
	res, err := c.rclient.R().
		SetBody(&joinRequest{
			ServerID: serverID,
			Addr:     addr,
		}).
		SetResult(&response{}).
		Post(c.urlPerfix + "/raft/join")
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform join request")

		return fmt.Errorf("failed to perform join request: %w", err)
	}

	if _, err := c.parseResponse(res); err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform join request")

		return err
	}

	return nil
}
