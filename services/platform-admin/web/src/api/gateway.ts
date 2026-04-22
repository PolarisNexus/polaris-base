import { rpc } from "../connect";

const SERVICE = "polaris.platform_admin.v1.GatewayService";

export interface Route {
  id: string;
  uri?: string;
  hosts?: string[];
  methods?: string[];
  upstreamId?: string;
  serviceId?: string;
  desc?: string;
  priority?: number;
  plugins?: Record<string, unknown>;
}

export interface PageInfo {
  page?: number;
  pageSize?: number;
  total?: number;
}

export interface ListRoutesResponse {
  items?: Route[];
  pageInfo?: PageInfo;
}

export function listRoutes(query = ""): Promise<ListRoutesResponse> {
  return rpc<{ query?: string }, ListRoutesResponse>(SERVICE, "ListRoutes", { query });
}
