// PKCE helpers — 手写避免引入 oidc-client-ts 等额外依赖。
// RFC 7636：code_verifier = 43-128 chars base64url，code_challenge = base64url(SHA256(verifier))。

export function randomString(length = 64): string {
  const buf = new Uint8Array(length);
  crypto.getRandomValues(buf);
  return base64url(buf);
}

export async function sha256(input: string): Promise<string> {
  const data = new TextEncoder().encode(input);
  const digest = await crypto.subtle.digest("SHA-256", data);
  return base64url(new Uint8Array(digest));
}

function base64url(buf: Uint8Array): string {
  let s = "";
  for (const b of buf) s += String.fromCharCode(b);
  return btoa(s).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}
