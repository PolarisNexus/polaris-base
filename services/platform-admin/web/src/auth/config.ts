// Vite build-time env 注入；运行时 fallback 指向开发默认值。
// issuer 需末尾带 `/`（Authentik OIDC 要求）。
export const authConfig = {
  issuer:
    import.meta.env.VITE_OIDC_ISSUER ??
    "http://localhost:9000/application/o/platform-admin/",
  clientId: import.meta.env.VITE_OIDC_CLIENT_ID ?? "polaris-platform-admin",
  redirectUri:
    import.meta.env.VITE_OIDC_REDIRECT_URI ??
    `${window.location.origin}/auth/callback`,
  scope: "openid email profile",
};
