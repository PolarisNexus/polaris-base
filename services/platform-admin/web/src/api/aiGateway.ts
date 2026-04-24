import { rpc } from "../connect";
import { PageInfo } from "./gateway";

const SERVICE = "polaris.platform_admin.v1.AiGatewayService";

export interface Provider {
  id: string;
  displayName: string;
  apisixProvider: string;
  baseUrl: string;
  supportedPaths: string[];
  status: string;
  upstreamEndpoint: string;
}

export interface UsageRecord {
  timestamp?: string;
  user?: string;
  provider?: string;
  model?: string;
  promptTokens?: number;
  completionTokens?: number;
  totalTokens?: number;
  statusCode?: number;
  latencyMs?: number;
  requestId?: string;
}

export interface UsageSummary {
  totalRequests?: number;
  totalPromptTokens?: number;
  totalCompletionTokens?: number;
  requestsByProvider?: Record<string, number>;
  requestsByUser?: Record<string, number>;
}

export interface QueryUsageReq {
  timeRange?: { start?: string; end?: string };
  user?: string;
  provider?: string;
  model?: string;
  page?: { page?: number; pageSize?: number };
  order?: "SORT_ORDER_ASC" | "SORT_ORDER_DESC";
}

export interface QueryUsageResp {
  items?: UsageRecord[];
  pageInfo?: PageInfo;
  summary?: UsageSummary;
}

export interface Quota {
  id: string;
  user: string;
  dailyTokenLimit: number;
  usedToday: number;
  window: string;
}

export interface ListQuotasResp {
  items?: Quota[];
  phaseNote?: string;
}

export const aiGatewayApi = {
  listProviders: () =>
    rpc<Record<string, never>, { items?: Provider[] }>(SERVICE, "ListProviders", {}),
  queryUsage: (r: QueryUsageReq) =>
    rpc<QueryUsageReq, QueryUsageResp>(SERVICE, "QueryUsage", r),
  listQuotas: () =>
    rpc<Record<string, never>, ListQuotasResp>(SERVICE, "ListQuotas", {}),
};
