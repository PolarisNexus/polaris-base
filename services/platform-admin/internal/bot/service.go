package bot

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/common/v1"
	v1 "github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1"
	"github.com/PolarisNexus/polaris-base/api/gen/go/polaris/platform_admin/v1/platform_adminv1connect"
)

type Service struct {
	platform_adminv1connect.UnimplementedBotServiceHandler
	client *Client
}

func NewService(client *Client) *Service { return &Service{client: client} }

func (s *Service) ListDecisions(ctx context.Context, req *connect.Request[v1.ListDecisionsRequest]) (*connect.Response[v1.ListDecisionsResponse], error) {
	q := url.Values{}
	if v := req.Msg.GetScope(); v != "" {
		q.Set("scope", v)
	}
	if v := req.Msg.GetValue(); v != "" {
		q.Set("value", v)
	}
	if v := req.Msg.GetOrigin(); v != "" {
		q.Set("origin", v)
	}
	// LAPI 默认返回全部未过期；active_only=false 时也只能返回未过期，历史决策需查 alerts。
	decisions, err := s.client.ListDecisions(ctx, q)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*v1.Decision, 0, len(decisions))
	for _, d := range decisions {
		items = append(items, decisionToProto(d))
	}
	return connect.NewResponse(&v1.ListDecisionsResponse{
		Items:    items,
		PageInfo: &commonv1.PageInfo{Total: int64(len(items))},
	}), nil
}

func (s *Service) CreateDecision(ctx context.Context, req *connect.Request[v1.CreateDecisionRequest]) (*connect.Response[v1.CreateDecisionResponse], error) {
	scope := req.Msg.GetScope()
	value := req.Msg.GetValue()
	if scope == "" || value == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("scope and value required"))
	}
	dType := req.Msg.GetType()
	if dType == "" {
		dType = "ban"
	}
	dur := "4h"
	if d := req.Msg.GetDuration(); d != nil {
		dur = d.AsDuration().String()
	}
	reason := req.Msg.GetReason()
	if reason == "" {
		reason = "manual action via platform-admin"
	}
	if _, err := s.client.CreateDecision(ctx, scope, value, dType, dur, reason); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// 回查刚创建的决策。
	q := url.Values{}
	q.Set("scope", scope)
	q.Set("value", value)
	decs, err := s.client.ListDecisions(ctx, q)
	if err != nil || len(decs) == 0 {
		return connect.NewResponse(&v1.CreateDecisionResponse{
			Decision: &v1.Decision{Scope: scope, Value: value, Type: dType, Origin: "platform-admin"},
		}), nil
	}
	return connect.NewResponse(&v1.CreateDecisionResponse{Decision: decisionToProto(decs[0])}), nil
}

func (s *Service) DeleteDecision(ctx context.Context, req *connect.Request[v1.DeleteDecisionRequest]) (*connect.Response[v1.DeleteDecisionResponse], error) {
	var (
		n   int
		err error
	)
	switch {
	case req.Msg.GetId() != 0:
		n, err = s.client.DeleteDecisionByID(ctx, req.Msg.GetId())
	case req.Msg.GetScope() != "" && req.Msg.GetValue() != "":
		n, err = s.client.DeleteDecisionByScopeValue(ctx, req.Msg.GetScope(), req.Msg.GetValue())
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id or scope+value required"))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.DeleteDecisionResponse{DeletedCount: int32(n)}), nil
}

func (s *Service) ListAlerts(ctx context.Context, req *connect.Request[v1.ListAlertsRequest]) (*connect.Response[v1.ListAlertsResponse], error) {
	q := url.Values{}
	if v := req.Msg.GetScenario(); v != "" {
		q.Set("scenario", v)
	}
	if v := req.Msg.GetSourceIp(); v != "" {
		q.Set("ip", v)
	}
	if tr := req.Msg.GetTimeRange(); tr != nil {
		if t := tr.GetStart(); t != nil {
			q.Set("since", t.AsTime().UTC().Format(time.RFC3339))
		}
	}
	if p := req.Msg.GetPage(); p != nil && p.GetPageSize() > 0 {
		q.Set("limit", strconv.Itoa(int(p.GetPageSize())))
	}
	alerts, err := s.client.ListAlerts(ctx, q)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*v1.Alert, 0, len(alerts))
	for _, a := range alerts {
		items = append(items, alertToProto(a))
	}
	return connect.NewResponse(&v1.ListAlertsResponse{
		Items:    items,
		PageInfo: &commonv1.PageInfo{Total: int64(len(items))},
	}), nil
}

func (s *Service) GetAlert(ctx context.Context, req *connect.Request[v1.GetAlertRequest]) (*connect.Response[v1.GetAlertResponse], error) {
	if req.Msg.GetId() == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id required"))
	}
	a, err := s.client.GetAlert(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&v1.GetAlertResponse{Alert: alertToProto(*a)}), nil
}

func decisionToProto(d Decision) *v1.Decision {
	p := &v1.Decision{
		Id:        d.ID,
		Origin:    d.Origin,
		Type:      d.Type,
		Scope:     d.Scope,
		Value:     d.Value,
		Scenario:  d.Scenario,
		Simulated: d.Simulated,
	}
	if d.Duration != "" {
		if dur, err := time.ParseDuration(d.Duration); err == nil {
			p.Duration = durationpb.New(dur)
			p.ExpiresAt = timestamppb.New(time.Now().Add(dur))
		}
	}
	return p
}

func alertToProto(a Alert) *v1.Alert {
	p := &v1.Alert{
		Id:           a.ID,
		Scenario:     a.Scenario,
		ScenarioHash: a.ScenarioHash,
		SourceIp:     a.Source.IP,
		SourceScope:  a.Source.Scope,
		EventsCount:  a.EventsCount,
		Message:      a.Message,
	}
	if t, err := time.Parse(time.RFC3339, a.StartAt); err == nil {
		p.StartedAt = timestamppb.New(t)
	}
	if t, err := time.Parse(time.RFC3339, a.StopAt); err == nil {
		p.StoppedAt = timestamppb.New(t)
	}
	for _, d := range a.Decisions {
		p.DecisionIds = append(p.DecisionIds, d.ID)
	}
	return p
}

