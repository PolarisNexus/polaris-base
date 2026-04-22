// 最小 Connect HTTP/JSON 客户端。Connect 协议 = POST /<Service>/<Method>，
// Content-Type: application/json，body = 请求消息 JSON。
// 错误响应 Content-Type: application/json，body = { code, message }。
// 参考：https://connectrpc.com/docs/protocol

import { getToken, logout } from "./auth/oidc";

export class ConnectError extends Error {
  constructor(public code: string, message: string) {
    super(message);
  }
}

export async function rpc<Req, Resp>(
  service: string,
  method: string,
  req: Req,
): Promise<Resp> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const resp = await fetch(`/${service}/${method}`, {
    method: "POST",
    headers,
    body: JSON.stringify(req),
  });
  if (resp.status === 401) {
    // token 失效 → 清本地状态并回登录
    void logout();
    throw new ConnectError("unauthenticated", "登录已失效，请重新登录");
  }
  const text = await resp.text();
  const data = text ? JSON.parse(text) : {};
  if (!resp.ok) {
    throw new ConnectError(data.code ?? "unknown", data.message ?? resp.statusText);
  }
  return data as Resp;
}
