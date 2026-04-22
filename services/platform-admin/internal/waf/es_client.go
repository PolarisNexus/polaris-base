package waf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ESClient 是 Elasticsearch _search 的最小客户端，仅用于 WAF 攻击日志查询。
type ESClient struct {
	baseURL  string
	username string
	password string
	http     *http.Client
}

func NewESClient(baseURL, username, password string) *ESClient {
	return &ESClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

type searchHit struct {
	Source json.RawMessage `json:"_source"`
	ID     string          `json:"_id"`
}

type searchResp struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []searchHit `json:"hits"`
	} `json:"hits"`
}

func (c *ESClient) Search(ctx context.Context, index string, query map[string]any) ([]json.RawMessage, int64, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/"+index+"/_search", bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		// index 尚未创建：视为空结果。
		return nil, 0, nil
	}
	if resp.StatusCode >= 400 {
		return nil, 0, fmt.Errorf("es search %s: %d %s", index, resp.StatusCode, raw)
	}
	var sr searchResp
	if err := json.Unmarshal(raw, &sr); err != nil {
		return nil, 0, err
	}
	out := make([]json.RawMessage, 0, len(sr.Hits.Hits))
	for _, h := range sr.Hits.Hits {
		out = append(out, h.Source)
	}
	return out, sr.Hits.Total.Value, nil
}
