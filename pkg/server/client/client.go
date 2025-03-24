package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/pkg/utils"
	"github.com/weastur/maf/pkg/utils/logging"
	restyzerolog "github.com/weastur/resty-zerolog"
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
	raftJoinPath                 = "/raft/join"
	raftKVPath                   = "/raft/kv"
	raftForgetPath               = "/raft/forget"
	raftInfoPath                 = "/raft/info"
)

type Client struct {
	Host      string
	urlPrefix string
	AuthToken string
	rclient   *resty.Client
	logger    zerolog.Logger
}

func New(host string, loggingEnabled bool) *Client {
	client := &Client{
		Host:      host,
		AuthToken: apiKey,
		rclient:   resty.New(),
		urlPrefix: host + "/api/v1alpha",
		logger:    log.With().Str(logging.ComponentCtxKey, "server-client").Logger(),
	}
	if !loggingEnabled {
		client.logger = client.logger.Level(zerolog.Disabled)
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
	client.rclient.SetLogger(restyzerolog.New(client.logger))
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

func NewWithTLS(host, serverCertFile string, loggingEnabled bool) *Client {
	client := New(host, loggingEnabled)

	client.rclient.SetRootCertificates(serverCertFile)

	return client
}

func NewWithMutualTLS(host, certFile, keyFile, serverCertFile string, loggingEnabled bool) *Client {
	client := New(host, loggingEnabled)

	client.rclient.SetRootCertificates(serverCertFile)
	client.rclient.SetCertificateFromFile(certFile, keyFile)

	return client
}

func NewWithAutoTLS(host string, config *TLSConfig, loggingEnabled bool) *Client {
	if config == nil || (config.CertFile == "" && config.KeyFile == "" && config.ServerCertFile == "") {
		return New(host, loggingEnabled)
	} else if config.CertFile == "" && config.KeyFile == "" {
		return NewWithTLS(host, config.ServerCertFile, loggingEnabled)
	}

	return NewWithMutualTLS(host, config.CertFile, config.KeyFile, config.ServerCertFile, loggingEnabled)
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

func (c *Client) parseRaftKVGetResponse(data any) (*raftKVGetResponse, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var raftResp raftKVGetResponse

	err = json.Unmarshal(dataBytes, &raftResp)
	if err != nil {
		return nil, err
	}

	return &raftResp, nil
}

func (c *Client) makeURL(elem ...string) string {
	baseURL, err := url.Parse(c.urlPrefix)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse base URL. Can't continue")
	}

	baseURL.Path = path.Join(baseURL.Path, path.Join(elem...))

	return baseURL.String()
}

func (c *Client) RaftJoin(serverID, addr string) error {
	res, err := c.rclient.R().
		SetBody(&raftJoinRequest{
			ServerID: serverID,
			Addr:     addr,
		}).
		SetResult(&response{}).
		Post(c.makeURL(raftJoinPath))
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

func (c *Client) RaftForget(serverID string) error {
	res, err := c.rclient.R().
		SetBody(&raftForgetRequest{
			ServerID: serverID,
		}).
		SetResult(&response{}).
		Post(c.makeURL(raftForgetPath))
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform forget request")

		return fmt.Errorf("failed to perform forget request: %w", err)
	}

	if _, err := c.parseResponse(res); err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform forget request")

		return err
	}

	return nil
}

func (c *Client) RaftKVGet(key string) (string, bool, error) {
	res, err := c.rclient.R().
		SetBody(&raftKVGetRequest{
			Key: key,
		}).
		SetResult(&response{}).
		Get(c.makeURL(raftKVPath, key))
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform KV get request")

		return "", false, fmt.Errorf("failed to perform KV get request: %w", err)
	}

	wrappedRes, err := c.parseResponse(res)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform KV get request")

		return "", false, err
	}

	kvData, err := c.parseRaftKVGetResponse(wrappedRes)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to parse KV get response")

		return "", false, ErrUnknownResponseFormat
	}

	return kvData.Value, kvData.Exist, nil
}

func (c *Client) RaftKVSet(key, value string) error {
	res, err := c.rclient.R().
		SetBody(&raftKVSetRequest{
			Key:   key,
			Value: value,
		}).
		SetResult(&response{}).
		Post(c.makeURL(raftKVPath))
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform KV set request")

		return fmt.Errorf("failed to perform KV set request: %w", err)
	}

	if _, err := c.parseResponse(res); err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform KV set request")

		return err
	}

	return nil
}

func (c *Client) RaftKVDelete(key string) error {
	res, err := c.rclient.R().
		SetResult(&response{}).
		Delete(c.makeURL(raftKVPath, key))
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform KV delete request")

		return fmt.Errorf("failed to perform KV delete request: %w", err)
	}

	if _, err := c.parseResponse(res); err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform KV delete request")

		return err
	}

	return nil
}

func (c *Client) RaftInfo(includeStats bool) (any, error) {
	res, err := c.rclient.R().
		SetQueryParam("include_stats", strconv.FormatBool(includeStats)).
		SetResult(&response{}).
		Get(c.makeURL(raftInfoPath))
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform raft info request")

		return nil, fmt.Errorf("failed to perform raft info request: %w", err)
	}

	data, err := c.parseResponse(res)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform raft info request")

		return nil, err
	}

	return data, nil
}
