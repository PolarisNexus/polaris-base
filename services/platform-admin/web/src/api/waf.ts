import { rpc } from "../connect";
import { PageInfo } from "./gateway";

const SERVICE = "polaris.platform_admin.v1.WafService";

export interface AttackLog {
  timestamp?: string;
  clientIp?: string;
  host?: string;
  uri?: string;
  method?: string;
  ruleId?: number;
  ruleMessage?: string;
  severity?: string;
  matchedData?: string;
  action?: string;
  routeId?: string;
  requestId?: string;
}

export interface QueryAttackLogsReq {
  timeRange?: { start?: string; end?: string };
  clientIp?: string;
  routeId?: string;
  ruleId?: number;
  severity?: string;
  query?: string;
  order?: "SORT_ORDER_ASC" | "SORT_ORDER_DESC";
  page?: { page?: number; pageSize?: number };
}

export interface QueryAttackLogsResp {
  items?: AttackLog[];
  pageInfo?: PageInfo;
}

export const wafApi = {
  queryAttackLogs: (r: QueryAttackLogsReq) =>
    rpc<QueryAttackLogsReq, QueryAttackLogsResp>(SERVICE, "QueryAttackLogs", r),
};
