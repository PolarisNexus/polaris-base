package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	commonv1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/common/v1"
	v1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1"
	"github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1/platform_adminv1connect"
)

type Service struct {
	platform_adminv1connect.UnimplementedGatewayServiceHandler
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client: client}
}

func (s *Service) ListRoutes(ctx context.Context, req *connect.Request[v1.ListRoutesRequest]) (*connect.Response[v1.ListRoutesResponse], error) {
	raws, total, err := s.client.List(ctx, "routes")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*v1.Route, 0, len(raws))
	for _, raw := range raws {
		r, err := decodeRoute(raw)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		items = append(items, r)
	}
	return connect.NewResponse(&v1.ListRoutesResponse{
		Items:    items,
		PageInfo: pageInfo(req.Msg.GetPage(), total),
	}), nil
}

func (s *Service) GetRoute(ctx context.Context, req *connect.Request[v1.GetRouteRequest]) (*connect.Response[v1.GetRouteResponse], error) {
	if req.Msg.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id required"))
	}
	raw, err := s.client.Get(ctx, "routes", req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	r, err := decodeRoute(raw)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.GetRouteResponse{Route: r}), nil
}

func (s *Service) ListUpstreams(ctx context.Context, req *connect.Request[v1.ListUpstreamsRequest]) (*connect.Response[v1.ListUpstreamsResponse], error) {
	raws, total, err := s.client.List(ctx, "upstreams")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*v1.Upstream, 0, len(raws))
	for _, raw := range raws {
		u, err := decodeUpstream(raw)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		items = append(items, u)
	}
	return connect.NewResponse(&v1.ListUpstreamsResponse{
		Items:    items,
		PageInfo: pageInfo(req.Msg.GetPage(), total),
	}), nil
}

func (s *Service) GetUpstream(ctx context.Context, req *connect.Request[v1.GetUpstreamRequest]) (*connect.Response[v1.GetUpstreamResponse], error) {
	if req.Msg.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id required"))
	}
	raw, err := s.client.Get(ctx, "upstreams", req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	u, err := decodeUpstream(raw)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.GetUpstreamResponse{Upstream: u}), nil
}

func (s *Service) UpdateRoutePlugins(ctx context.Context, req *connect.Request[v1.UpdateRoutePluginsRequest]) (*connect.Response[v1.UpdateRoutePluginsResponse], error) {
	id := req.Msg.GetRouteId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route_id required"))
	}
	plugins := map[string]any{}
	for k, v := range req.Msg.GetPlugins() {
		plugins[k] = v.AsMap()
	}
	body, err := json.Marshal(map[string]any{"plugins": plugins})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	raw, err := s.client.Patch(ctx, "routes", id, body)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	r, err := decodeRoute(raw)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.UpdateRoutePluginsResponse{Route: r}), nil
}

func (s *Service) SetRouteRateLimit(ctx context.Context, req *connect.Request[v1.SetRouteRateLimitRequest]) (*connect.Response[v1.SetRouteRateLimitResponse], error) {
	id := req.Msg.GetRouteId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("route_id required"))
	}
	var patch map[string]any
	if req.Msg.GetDisable() {
		patch = map[string]any{"plugins": map[string]any{"limit-count": nil}}
	} else {
		key := req.Msg.GetKey()
		if key == "" {
			key = "remote_addr"
		}
		patch = map[string]any{
			"plugins": map[string]any{
				"limit-count": map[string]any{
					"count":        req.Msg.GetCount(),
					"time_window":  req.Msg.GetWindowSeconds(),
					"key_type":     "var",
					"key":          key,
					"rejected_code": 429,
				},
			},
		}
	}
	body, err := json.Marshal(patch)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	raw, err := s.client.Patch(ctx, "routes", id, body)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	r, err := decodeRoute(raw)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.SetRouteRateLimitResponse{Route: r}), nil
}

func decodeRoute(raw json.RawMessage) (*v1.Route, error) {
	var a apisixRoute
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("decode route: %w", err)
	}
	plugins := map[string]*structpb.Struct{}
	for k, v := range a.Plugins {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		s, err := structpb.NewStruct(m)
		if err == nil {
			plugins[k] = s
		}
	}
	return &v1.Route{
		Id:         a.ID,
		Uri:        a.URI,
		Hosts:      a.Hosts,
		Methods:    a.Methods,
		UpstreamId: a.UpstreamID,
		ServiceId:  a.ServiceID,
		Desc:       a.Desc,
		Priority:   a.Priority,
		CreateTime: a.CreateTime,
		UpdateTime: a.UpdateTime,
		Plugins:    plugins,
	}, nil
}

func decodeUpstream(raw json.RawMessage) (*v1.Upstream, error) {
	var a apisixUpstream
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("decode upstream: %w", err)
	}
	return &v1.Upstream{
		Id:               a.ID,
		Type:             a.Type,
		Nodes:            a.Nodes,
		Scheme:           a.Scheme,
		TimeoutConnectMs: int32(a.Timeout.Connect * 1000),
		TimeoutSendMs:    int32(a.Timeout.Send * 1000),
		TimeoutReadMs:    int32(a.Timeout.Read * 1000),
		Desc:             a.Desc,
	}, nil
}

type apisixRoute struct {
	ID         string         `json:"id"`
	URI        string         `json:"uri"`
	Hosts      []string       `json:"hosts"`
	Methods    []string       `json:"methods"`
	UpstreamID string         `json:"upstream_id"`
	ServiceID  string         `json:"service_id"`
	Desc       string         `json:"desc"`
	Priority   int32          `json:"priority"`
	CreateTime int64          `json:"create_time"`
	UpdateTime int64          `json:"update_time"`
	Plugins    map[string]any `json:"plugins"`
}

type apisixUpstream struct {
	ID     string           `json:"id"`
	Type   string           `json:"type"`
	Nodes  map[string]int32 `json:"nodes"`
	Scheme string           `json:"scheme"`
	Desc   string           `json:"desc"`
	Timeout struct {
		Connect float64 `json:"connect"`
		Send    float64 `json:"send"`
		Read    float64 `json:"read"`
	} `json:"timeout"`
}

func pageInfo(req *commonv1.PageRequest, total int64) *commonv1.PageInfo {
	var page, size int32
	if req != nil {
		page = req.GetPage()
		size = req.GetPageSize()
	}
	return &commonv1.PageInfo{Page: page, PageSize: size, Total: total}
}
