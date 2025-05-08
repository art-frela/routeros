package routeros

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/art-frela/routeros/types"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/time/rate"
)

var (
	errMissBaseURL    = errors.New("miss baseURL")
	errMissUserOrPass = errors.New("miss user/password")
)

type Config struct {
	BaseURL              string        `envconfig:"base_url" validate:"required"`
	RequestTimeout       time.Duration `envconfig:"request_timeout" default:"10s"`
	PauseBetweenRequests time.Duration `envconfig:"pause_between_requests" default:"100ms"`
	BurstRequestCount    int           `envconfig:"burst_req_count" default:"10"`
	User                 string        `envconfig:"user" default:"root"`
	Password             string        `envconfig:"password" default:"master"`
}

func NewClientConfigFromEnv(envPrefix string) (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(envPrefix, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

type Client struct {
	IPFirewallAddressListService *IPFirewallAddressListService
	ToolService                  *ToolService
	//
	baseURL        *url.URL
	lim            *rate.Limiter
	requestTimeout time.Duration
	user           string
	password       string
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errMissBaseURL
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse baseURL: %w", err)
	}

	if cfg.User == "" || cfg.Password == "" {
		return nil, errMissUserOrPass
	}

	u.Path = path.Join(u.Path, types.EndpointRest)

	rt := rate.Every(cfg.PauseBetweenRequests)

	c := &Client{
		baseURL:        u,
		lim:            rate.NewLimiter(rt, cfg.BurstRequestCount),
		requestTimeout: cfg.RequestTimeout,
		user:           cfg.User,
		password:       cfg.Password,
	}

	c.IPFirewallAddressListService = &IPFirewallAddressListService{c: c}
	c.ToolService = &ToolService{c: c}

	return c, nil
}

func makeRequest[T any](ctx context.Context, c *Client, apiEndpoint, method string, requestData any, queries url.Values) (T, error) {
	var sRes T

	if err := c.lim.Wait(ctx); err != nil {
		return sRes, err
	}

	var requestBody io.Reader

	if requestData != nil {
		body, err := json.Marshal(requestData)
		if err != nil {
			return sRes, err
		}

		requestBody = bytes.NewBuffer(body)
	}

	u := *c.baseURL.JoinPath(apiEndpoint)

	qq := u.Query()

	for k, vv := range queries {
		for _, v := range vv {
			qq.Add(k, v)
		}
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, method, u.String()+"?"+qq.Encode(), requestBody)
	if err != nil {
		return sRes, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.password)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sRes, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return sRes, err
		}

		return sRes, fmt.Errorf("status_code: %d, response: %s", res.StatusCode, string(body))
	}

	if err := json.NewDecoder(res.Body).Decode(&sRes); err != nil {
		return sRes, err
	}

	return sRes, nil
}
