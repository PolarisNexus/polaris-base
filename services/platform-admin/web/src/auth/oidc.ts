import { authConfig } from "./config";
import { randomString, sha256 } from "./pkce";

type Discovery = {
  authorization_endpoint: string;
  token_endpoint: string;
  end_session_endpoint?: string;
};

const TOKEN_KEY = "polaris.idToken";
const EXPIRES_KEY = "polaris.idTokenExp";
const PKCE_VERIFIER_KEY = "polaris.pkceVerifier";
const PKCE_STATE_KEY = "polaris.pkceState";
const RETURN_TO_KEY = "polaris.returnTo";

let discoveryPromise: Promise<Discovery> | null = null;

function discover(): Promise<Discovery> {
  if (!discoveryPromise) {
    const url = authConfig.issuer.replace(/\/$/, "") + "/.well-known/openid-configuration";
    discoveryPromise = fetch(url).then((r) => {
      if (!r.ok) throw new Error(`oidc discovery failed: ${r.status}`);
      return r.json() as Promise<Discovery>;
    });
  }
  return discoveryPromise;
}

export async function login(returnTo = window.location.pathname + window.location.search) {
  const meta = await discover();
  const verifier = randomString(64);
  const challenge = await sha256(verifier);
  const state = randomString(24);
  sessionStorage.setItem(PKCE_VERIFIER_KEY, verifier);
  sessionStorage.setItem(PKCE_STATE_KEY, state);
  sessionStorage.setItem(RETURN_TO_KEY, returnTo);

  const params = new URLSearchParams({
    response_type: "code",
    client_id: authConfig.clientId,
    redirect_uri: authConfig.redirectUri,
    scope: authConfig.scope,
    state,
    code_challenge: challenge,
    code_challenge_method: "S256",
  });
  window.location.assign(`${meta.authorization_endpoint}?${params}`);
}

export async function handleCallback(): Promise<string> {
  const url = new URL(window.location.href);
  const code = url.searchParams.get("code");
  const state = url.searchParams.get("state");
  const expectedState = sessionStorage.getItem(PKCE_STATE_KEY);
  const verifier = sessionStorage.getItem(PKCE_VERIFIER_KEY);
  if (!code || !state || state !== expectedState || !verifier) {
    throw new Error("invalid oidc callback");
  }
  const meta = await discover();
  const body = new URLSearchParams({
    grant_type: "authorization_code",
    code,
    redirect_uri: authConfig.redirectUri,
    client_id: authConfig.clientId,
    code_verifier: verifier,
  });
  const resp = await fetch(meta.token_endpoint, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
  });
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(`token exchange failed: ${resp.status} ${text}`);
  }
  const data = (await resp.json()) as { id_token: string; expires_in: number };
  const exp = Date.now() + (data.expires_in ?? 3600) * 1000;
  localStorage.setItem(TOKEN_KEY, data.id_token);
  localStorage.setItem(EXPIRES_KEY, String(exp));
  sessionStorage.removeItem(PKCE_VERIFIER_KEY);
  sessionStorage.removeItem(PKCE_STATE_KEY);
  const returnTo = sessionStorage.getItem(RETURN_TO_KEY) ?? "/";
  sessionStorage.removeItem(RETURN_TO_KEY);
  return returnTo;
}

export function getToken(): string | null {
  const token = localStorage.getItem(TOKEN_KEY);
  const exp = Number(localStorage.getItem(EXPIRES_KEY) ?? 0);
  if (!token || Date.now() >= exp) {
    return null;
  }
  return token;
}

export async function logout() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(EXPIRES_KEY);
  try {
    const meta = await discover();
    if (meta.end_session_endpoint) {
      window.location.assign(meta.end_session_endpoint);
      return;
    }
  } catch {
    /* fall through */
  }
  window.location.assign("/");
}

// IDToken payload 解码（不校验签名——校验由 BFF 完成；前端仅用于展示用户名）。
export function decodeToken(token: string): Record<string, unknown> | null {
  const parts = token.split(".");
  if (parts.length !== 3) return null;
  try {
    const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const json = atob(payload.padEnd(payload.length + ((4 - (payload.length % 4)) % 4), "="));
    return JSON.parse(json);
  } catch {
    return null;
  }
}
