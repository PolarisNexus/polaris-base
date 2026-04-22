import { rpc } from "../connect";
import { PageInfo } from "./gateway";

const SERVICE = "polaris.platform_admin.v1.BotService";

export interface Decision {
  id: string; // int64 as string in Connect JSON
  origin?: string;
  type?: string;
  scope?: string;
  value?: string;
  scenario?: string;
  duration?: string; // google.protobuf.Duration JSON form: "3600s"
  createdAt?: string;
  expiresAt?: string;
  simulated?: boolean;
}

export interface Alert {
  id: string;
  scenario?: string;
  scenarioHash?: string;
  sourceIp?: string;
  sourceScope?: string;
  eventsCount?: number;
  startedAt?: string;
  stoppedAt?: string;
  decisionIds?: string[];
  message?: string;
}

export interface ListDecisionsReq {
  scope?: string;
  value?: string;
  origin?: string;
  activeOnly?: boolean;
}

export interface ListDecisionsResp {
  items?: Decision[];
  pageInfo?: PageInfo;
}

export interface CreateDecisionReq {
  scope: string;
  value: string;
  type?: string;
  duration?: string; // "4h", "30m" → Connect JSON 支持秒后缀："14400s"
  reason: string;
}

export interface DeleteDecisionReq {
  id?: string;
  scope?: string;
  value?: string;
  reason?: string;
}

export interface ListAlertsReq {
  sourceIp?: string;
  scenario?: string;
  timeRange?: { start?: string; end?: string };
  page?: { page?: number; pageSize?: number };
}

export interface ListAlertsResp {
  items?: Alert[];
  pageInfo?: PageInfo;
}

export const botApi = {
  listDecisions: (r: ListDecisionsReq) => rpc<ListDecisionsReq, ListDecisionsResp>(SERVICE, "ListDecisions", r),
  createDecision: (r: CreateDecisionReq) => rpc<CreateDecisionReq, { decision?: Decision }>(SERVICE, "CreateDecision", r),
  deleteDecision: (r: DeleteDecisionReq) => rpc<DeleteDecisionReq, { deletedCount?: number }>(SERVICE, "DeleteDecision", r),
  listAlerts: (r: ListAlertsReq) => rpc<ListAlertsReq, ListAlertsResp>(SERVICE, "ListAlerts", r),
  getAlert: (id: string) => rpc<{ id: string }, { alert?: Alert }>(SERVICE, "GetAlert", { id }),
};
