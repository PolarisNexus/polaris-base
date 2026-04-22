import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { decodeToken, getToken, login, logout } from "./oidc";

type User = { sub: string; email?: string; name?: string };

type AuthState = {
  token: string | null;
  user: User | null;
  login: () => void;
  logout: () => void;
};

const Ctx = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => getToken());

  // 浏览器 tab 之间同步登出
  useEffect(() => {
    const h = () => setToken(getToken());
    window.addEventListener("storage", h);
    return () => window.removeEventListener("storage", h);
  }, []);

  const user = useMemo<User | null>(() => {
    if (!token) return null;
    const c = decodeToken(token) ?? {};
    return {
      sub: (c.sub as string) ?? "",
      email: c.email as string | undefined,
      name: (c.preferred_username as string) ?? (c.name as string) ?? (c.email as string),
    };
  }, [token]);

  const value: AuthState = {
    token,
    user,
    login: () => void login(),
    logout: () => void logout(),
  };
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useAuth(): AuthState {
  const v = useContext(Ctx);
  if (!v) throw new Error("useAuth must be used within AuthProvider");
  return v;
}
