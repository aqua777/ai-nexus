package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"log/slog"
)

const (
	DefaultTimeout  = 30 * time.Second
	DefaultScheme   = "http"
	DefaultHost     = "127.0.0.1"
	SchemeSeparator = "://"
)

type Client struct {
	baseUrl    string
	timeout    time.Duration
	clientOnce sync.Once
	client     *http.Client
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	if c.client != nil {
		return c
	}
	c.timeout = timeout
	return c
}

func (c *Client) getClient() *http.Client {
	c.clientOnce.Do(func() {
		c.client = &http.Client{
			Timeout: c.timeout,
		}
	})
	return c.client
}

func (c *Client) getFullUrl(path string) string {
	return c.baseUrl + strings.ReplaceAll(path, "//", "/")
}

func (c *Client) Do(ctx context.Context, method, path string, headers map[string]string, dataBytes []byte) (data []byte, status int, err error) {
	slog.Debug("HttpClient.Do()", "method", method, "path", path, "headers", headers, "dataBytes", string(dataBytes))
	req, err := http.NewRequestWithContext(ctx, method, c.getFullUrl(path), bytes.NewReader(dataBytes))
	if err != nil {
		return nil, 0, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.getClient().Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	slog.Debug("HttpClient.Do()", "respBody", string(respBody), "statusCode", resp.StatusCode)

	return respBody, resp.StatusCode, nil
}

func NewClient(optionalBaseUrl ...string) (*Client, error) {
	var baseUrl string
	if len(optionalBaseUrl) == 1 {
		url, err := getBaseUrl(optionalBaseUrl[0])
		if err != nil {
			return nil, err
		}
		baseUrl = url
	} else {
		baseUrl = ""
	}
	return &Client{
		timeout: DefaultTimeout,
		baseUrl: baseUrl,
	}, nil
}
