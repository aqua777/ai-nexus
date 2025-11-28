package http

import (
	"context"
	"encoding/json"
	"fmt"
)

type JsonClient struct {
	Client *Client
}

func (c *JsonClient) Do(ctx context.Context, method, path string, reqObj, respObj any, headers map[string]string) (err error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[ContentTypeHeader] = ContentTypeJson
	reqData, err := json.Marshal(reqObj)
	if err != nil {
		return err
	}
	respBytes, status, err := c.Client.Do(ctx, method, path, headers, reqData)
	if err != nil {
		return err
	} else if status != StatusOK {
		// Check if response body contains error message
		if len(respBytes) > 0 {
			var errResp map[string]interface{}
			if jsonErr := json.Unmarshal(respBytes, &errResp); jsonErr == nil {
				if errMsg, ok := errResp["error"].(string); ok {
					return fmt.Errorf("status code: %d, error: %s", status, errMsg)
				}
			}
			// If not JSON error or couldn't parse, return raw body as string
			return fmt.Errorf("status code: %d, body: %s", status, string(respBytes))
		}
		return fmt.Errorf("status code: %d", status)
	}
	if respObj != nil {
		err = json.Unmarshal(respBytes, respObj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *JsonClient) Get(ctx context.Context, path string, respObj any, headers map[string]string) (err error) {
	return c.Do(ctx, MethodGet, path, nil, respObj, headers)
}

func (c *JsonClient) Post(ctx context.Context, path string, reqObj, respObj any, headers map[string]string) (err error) {
	return c.Do(ctx, MethodPost, path, reqObj, respObj, headers)
}

func (c *JsonClient) Put(ctx context.Context, path string, reqObj, respObj any, headers map[string]string) (err error) {
	return c.Do(ctx, MethodPut, path, reqObj, respObj, headers)
}

func (c *JsonClient) Delete(ctx context.Context, path string, reqObj, respObj any, headers map[string]string) (err error) {
	return c.Do(ctx, MethodDelete, path, reqObj, respObj, headers)
}

func (c *JsonClient) Patch(ctx context.Context, path string, reqObj, respObj any, headers map[string]string) (err error) {
	return c.Do(ctx, MethodPatch, path, reqObj, respObj, headers)
}

func (c *JsonClient) Options(ctx context.Context, path string, reqObj, respObj any, headers map[string]string) (err error) {
	return c.Do(ctx, MethodOptions, path, reqObj, respObj, headers)
}

func (c *JsonClient) Head(ctx context.Context, path string, reqObj, respObj any, headers map[string]string) (err error) {
	return c.Do(ctx, MethodHead, path, reqObj, respObj, headers)
}

func NewJsonClient(optionalBaseUrl ...string) (*JsonClient, error) {
	client, err := NewClient(optionalBaseUrl...)
	if err != nil {
		return nil, err
	}
	return &JsonClient{Client: client}, nil
}
