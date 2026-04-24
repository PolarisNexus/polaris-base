// Package ai_gateway 实现 AiGatewayService（ADR-0014 Phase I MVP）。
//
// Providers 走静态表；Usage 从 ES `ai-usage` index 查；Quotas 返回空占位（Phase II）。
package ai_gateway

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/common/v1"
	v1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1"
	"github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1/platform_adminv1connect"
)

// indexName 默认使用单一 index；生产挂 ES ILM 按天滚动。
const indexName = "ai-usage"

// ESSearcher 抽出最小依赖，避免与 waf.ESClient 强耦合。
// 现在两边各自一份实现，后续如果 ES 调用增多可提到 internal/es 共享包。
type ESSearcher interface {
	Search(ctx context.Context, index string, query map[string]any) ([]json.RawMessage, int64, error)
}

type Service struct {
	platform_adminv1connect.UnimplementedAiGatewayServiceHandler
	es ESSearcher
}

func NewService(es ESSearcher) *Service { return &Service{es: es} }

func (s *Service) ListProviders(ctx context.Context, req *connect.Request[v1.ListProvidersRequest]) (*connect.Response[v1.ListProvidersResponse], error) {
	return connect.NewResponse(&v1.ListProvidersResponse{Items: staticProviders()}), nil
}

func (s *Service) ListQuotas(ctx context.Context, req *connect.Request[v1.ListQuotasRequest]) (*connect.Response[v1.ListQuotasResponse], error) {
	return connect.NewResponse(&v1.ListQuotasResponse{
		Items:     nil,
		PhaseNote: "Quotas 在 ADR-0014 Phase II 实现；当前 APISIX ai-rate-limiting 未挂载。",
	}), nil
}

// QueryUsage 查 ES ai-usage index。文档结构见 elasticsearch-logger 默认 schema：
//
//	start_time (ms epoch), client_ip, route_id,
//	request.{uri, body, headers.{x-userinfo, x-request-id}},
//	response.{status, body}
//
// 解析 response.body（JSON 字符串）拿 usage.{prompt,completion}_tokens。
// user 字段：优先 X-Userinfo base64 解码出的 sub，退化到 client_ip。
func (s *Service) QueryUsage(ctx context.Context, req *connect.Request[v1.QueryUsageRequest]) (*connect.Response[v1.QueryUsageResponse], error) {
	page := req.Msg.GetPage()
	size := int32(50)
	from := int32(0)
	if page != nil {
		if p := page.GetPageSize(); p > 0 {
			size = p
		}
		if p := page.GetPage(); p > 1 {
			from = (p - 1) * size
		}
	}

	must := []map[string]any{}
	// URL 匹配 /ai/v1/<provider>/* 的流量
	must = append(must, map[string]any{
		"prefix": map[string]any{"request.uri": "/ai/v1/"},
	})
	if tr := req.Msg.GetTimeRange(); tr != nil {
		r := map[string]any{}
		if t := tr.GetStart(); t != nil {
			r["gte"] = t.AsTime().UnixMilli()
		}
		if t := tr.GetEnd(); t != nil {
			r["lte"] = t.AsTime().UnixMilli()
		}
		if len(r) > 0 {
			must = append(must, map[string]any{"range": map[string]any{"start_time": r}})
		}
	}
	if p := req.Msg.GetProvider(); p != "" {
		must = append(must, map[string]any{"prefix": map[string]any{"request.uri": "/ai/v1/" + p + "/"}})
	}

	sortOrder := "desc"
	if req.Msg.GetOrder() == commonv1.SortOrder_SORT_ORDER_ASC {
		sortOrder = "asc"
	}
	body := map[string]any{
		"from":  from,
		"size":  size,
		"query": map[string]any{"bool": map[string]any{"must": must}},
		"sort":  []map[string]any{{"start_time": map[string]any{"order": sortOrder}}},
	}

	sources, total, err := s.es.Search(ctx, indexName, body)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*v1.UsageRecord, 0, len(sources))
	summary := &v1.UsageSummary{
		RequestsByProvider: map[string]int64{},
		RequestsByUser:     map[string]int64{},
	}
	for _, src := range sources {
		rec := decodeUsage(src)
		// 应用 user / model 过滤（ES-side 可再优化，这里先简单过滤）
		if u := req.Msg.GetUser(); u != "" && rec.GetUser() != u {
			continue
		}
		if m := req.Msg.GetModel(); m != "" && rec.GetModel() != m {
			continue
		}
		items = append(items, rec)
		summary.TotalRequests++
		summary.TotalPromptTokens += int64(rec.GetPromptTokens())
		summary.TotalCompletionTokens += int64(rec.GetCompletionTokens())
		if p := rec.GetProvider(); p != "" {
			summary.RequestsByProvider[p]++
		}
		if u := rec.GetUser(); u != "" {
			summary.RequestsByUser[u]++
		}
	}

	return connect.NewResponse(&v1.QueryUsageResponse{
		Items:    items,
		PageInfo: &commonv1.PageInfo{Page: page.GetPage(), PageSize: size, Total: total},
		Summary:  summary,
	}), nil
}

// esUsageDoc 对齐 APISIX elasticsearch-logger 写入的嵌套结构。
// 只取我们关心的字段，未列出的 APISIX 扩展字段忽略。
type esUsageDoc struct {
	StartTime int64  `json:"start_time"`
	ClientIP  string `json:"client_ip"`
	Request   struct {
		URI     string `json:"uri"`
		Headers struct {
			XUserinfo string `json:"x-userinfo"`
			RequestID string `json:"x-request-id"`
		} `json:"headers"`
		// body 是 JSON 字符串；MVP 不直接解析客户端请求的 model，
		// 改从 response.body.model 回读（更准，代表 provider 实际用的 model）。
		Body string `json:"body"`
	} `json:"request"`
	Response struct {
		Status int32  `json:"status"`
		Body   string `json:"body"`
	} `json:"response"`
	Latency float64 `json:"latency"`
}

// userinfoClaims 取 JWT 的 sub 一字段。
// openid-connect 插件以 base64(json) 注入 X-Userinfo 到上游（logger 收到）。
type userinfoClaims struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
}

// respModelUsage 解析 response.body 的 OpenAI schema 子集。
type respModelUsage struct {
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
		TotalTokens      int32 `json:"total_tokens"`
	} `json:"usage"`
}

// providerFromURI 从 /ai/v1/<provider>/... 取 provider id。
var providerRe = regexp.MustCompile(`^/ai/v1/([^/]+)/`)

func decodeUsage(src json.RawMessage) *v1.UsageRecord {
	var doc esUsageDoc
	_ = json.Unmarshal(src, &doc)

	rec := &v1.UsageRecord{
		StatusCode: doc.Response.Status,
		LatencyMs:  doc.Latency,
		RequestId:  doc.Request.Headers.RequestID,
	}
	if doc.StartTime > 0 {
		rec.Timestamp = timestamppb.New(time.UnixMilli(doc.StartTime))
	}
	if m := providerRe.FindStringSubmatch(doc.Request.URI); len(m) == 2 {
		rec.Provider = m[1]
	}
	// user：优先 JWT sub
	if ui := doc.Request.Headers.XUserinfo; ui != "" {
		if raw, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(ui, "=")); err == nil {
			var c userinfoClaims
			if json.Unmarshal(raw, &c) == nil {
				if c.PreferredUsername != "" {
					rec.User = c.PreferredUsername
				} else if c.Sub != "" {
					rec.User = c.Sub
				}
			}
		}
	}
	if rec.User == "" {
		rec.User = doc.ClientIP
	}
	// model + usage 从响应体取
	if doc.Response.Body != "" {
		var r respModelUsage
		if json.Unmarshal([]byte(doc.Response.Body), &r) == nil {
			rec.Model = r.Model
			rec.PromptTokens = r.Usage.PromptTokens
			rec.CompletionTokens = r.Usage.CompletionTokens
			rec.TotalTokens = r.Usage.TotalTokens
		}
	}
	return rec
}
