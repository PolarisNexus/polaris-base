package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 是 APISIX Admin API 的最小客户端。
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// listResponse 是 APISIX Admin API list 响应的通用封装。
type listResponse struct {
	List []struct {
		Value json.RawMessage `json:"value"`
	} `json:"list"`
	Total int64 `json:"total"`
}

type getResponse struct {
	Value json.RawMessage `json:"value"`
}

func (c *Client) List(ctx context.Context, resource string) ([]json.RawMessage, int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/apisix/admin/"+resource, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, 0, fmt.Errorf("apisix admin %s: %d %s", resource, resp.StatusCode, body)
	}
	var lr listResponse
	if err := json.Unmarshal(body, &lr); err != nil {
		return nil, 0, err
	}
	items := make([]json.RawMessage, 0, len(lr.List))
	for _, it := range lr.List {
		items = append(items, it.Value)
	}
	return items, lr.Total, nil
}

func (c *Client) Get(ctx context.Context, resource, id string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/apisix/admin/"+resource+"/"+id, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("not found: %s/%s", resource, id)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("apisix admin %s/%s: %d %s", resource, id, resp.StatusCode, body)
	}
	var gr getResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		return nil, err
	}
	return gr.Value, nil
}

func (c *Client) Patch(ctx context.Context, resource, id string, body []byte) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "PATCH", c.baseURL+"/apisix/admin/"+resource+"/"+id, bytesReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("apisix admin patch %s/%s: %d %s", resource, id, resp.StatusCode, respBody)
	}
	var gr getResponse
	if err := json.Unmarshal(respBody, &gr); err != nil {
		return nil, err
	}
	return gr.Value, nil
}
