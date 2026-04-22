package waf

import (
	"context"
	"encoding/json"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/common/v1"
	v1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1"
	"github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1/platform_adminv1connect"
)

// indexName 默认使用单一 index；生产环境挂 ES ILM 滚动。
const indexName = "apisix-access"

type Service struct {
	platform_adminv1connect.UnimplementedWafServiceHandler
	es *ESClient
}

func NewService(es *ESClient) *Service { return &Service{es: es} }

func (s *Service) QueryAttackLogs(ctx context.Context, req *connect.Request[v1.QueryAttackLogsRequest]) (*connect.Response[v1.QueryAttackLogsResponse], error) {
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

	// MVP 口径：response.status >= 400 视为潜在攻击；精确区分待 Coraza 扩展响应头后接入。
	// APISIX elasticsearch-logger 写入字段为嵌套：request.uri/method、response.status、client_ip（flat）、start_time（epoch ms）。
	must = append(must, map[string]any{
		"range": map[string]any{"response.status": map[string]any{"gte": 400}},
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
	if ip := req.Msg.GetClientIp(); ip != "" {
		must = append(must, map[string]any{"term": map[string]any{"client_ip": ip}})
	}
	if rid := req.Msg.GetRouteId(); rid != "" {
		must = append(must, map[string]any{"term": map[string]any{"route_id": rid}})
	}
	if sev := req.Msg.GetSeverity(); sev != "" {
		must = append(must, map[string]any{"term": map[string]any{"severity": sev}})
	}
	if q := req.Msg.GetQuery(); q != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"request.uri", "request.url", "request.headers.host"},
			},
		})
	}

	sortOrder := "desc"
	if req.Msg.GetOrder() == commonv1.SortOrder_SORT_ORDER_ASC {
		sortOrder = "asc"
	}

	body := map[string]any{
		"from": from,
		"size": size,
		"query": map[string]any{
			"bool": map[string]any{"must": must},
		},
		"sort": []map[string]any{
			{"start_time": map[string]any{"order": sortOrder}},
		},
	}

	sources, total, err := s.es.Search(ctx, indexName, body)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*v1.AttackLog, 0, len(sources))
	for _, src := range sources {
		items = append(items, decodeAttackLog(src))
	}
	return connect.NewResponse(&v1.QueryAttackLogsResponse{
		Items:    items,
		PageInfo: &commonv1.PageInfo{Page: page.GetPage(), PageSize: size, Total: total},
	}), nil
}

// esAccessLog 覆盖 APISIX elasticsearch-logger 默认字段（嵌套结构）。
// 未命中的 proto 字段（rule_id / rule_message / severity / matched_data）
// 需要 coraza-proxy-wasm 后续通过 response header 回传或在 ES 侧做 ingest pipeline 提取。
type esAccessLog struct {
	StartTime int64 `json:"start_time"` // epoch ms
	ClientIP  string `json:"client_ip"`
	RouteID   string `json:"route_id"`
	Request   struct {
		URI     string `json:"uri"`
		URL     string `json:"url"`
		Method  string `json:"method"`
		Headers struct {
			Host      string `json:"host"`
			RequestID string `json:"x-request-id"`
		} `json:"headers"`
	} `json:"request"`
	Response struct {
		Status int32 `json:"status"`
	} `json:"response"`
}

func decodeAttackLog(src json.RawMessage) *v1.AttackLog {
	var a esAccessLog
	_ = json.Unmarshal(src, &a)
	log := &v1.AttackLog{
		ClientIp:  a.ClientIP,
		Host:      a.Request.Headers.Host,
		Uri:       firstNonEmpty(a.Request.URI, a.Request.URL),
		Method:    a.Request.Method,
		RouteId:   a.RouteID,
		RequestId: a.Request.Headers.RequestID,
		Action:    actionFromStatus(a.Response.Status),
	}
	if a.StartTime > 0 {
		log.Timestamp = timestamppb.New(time.UnixMilli(a.StartTime))
	}
	return log
}

func actionFromStatus(s int32) string {
	switch {
	case s == 403:
		return "block"
	case s >= 500:
		return "error"
	case s >= 400:
		return "reject"
	default:
		return "log"
	}
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

