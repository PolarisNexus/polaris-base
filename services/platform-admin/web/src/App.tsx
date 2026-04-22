import { Button, Dropdown, Layout, Menu } from "antd";
import { Link, Route, Routes, Navigate, useLocation } from "react-router-dom";
import RoutesPage from "./pages/Routes";
import BotDecisionsPage from "./pages/BotDecisions";
import BotAlertsPage from "./pages/BotAlerts";
import WafAttackLogsPage from "./pages/WafAttackLogs";
import AuthCallback from "./pages/AuthCallback";
import LoginPage from "./pages/Login";
import { useAuth } from "./auth/AuthContext";

function Shell() {
  const loc = useLocation();
  const selected = loc.pathname.split("/")[1] || "routes";
  const { user, logout } = useAuth();
  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Layout.Sider width={220}>
        <div style={{ color: "#fff", padding: "16px 24px", fontWeight: 600 }}>
          Platform Admin
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selected]}
          items={[
            { key: "routes", label: <Link to="/routes">网关路由</Link> },
            {
              key: "waf",
              label: "WAF",
              children: [
                { key: "waf-attacks", label: <Link to="/waf-attacks">攻击日志</Link> },
                { key: "bot-decisions", label: <Link to="/bot-decisions">Bot 决策</Link> },
                { key: "bot-alerts", label: <Link to="/bot-alerts">Bot 告警</Link> },
                { key: "waf-rules", label: <span style={{ opacity: 0.5 }}>规则 · 下一阶段</span>, disabled: true },
              ],
            },
            {
              key: "iam",
              label: "IAM",
              children: [
                {
                  key: "iam-authentik",
                  label: (
                    <a
                      href="http://localhost:9000/if/admin/"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      Authentik Admin ↗
                    </a>
                  ),
                },
              ],
            },
          ]}
        />
      </Layout.Sider>
      <Layout>
        <Layout.Header style={{ background: "#fff", display: "flex", justifyContent: "flex-end", paddingRight: 24 }}>
          <Dropdown
            menu={{
              items: [{ key: "logout", label: "登出", onClick: () => logout() }],
            }}
          >
            <Button type="text">{user?.name ?? user?.sub ?? "未知用户"}</Button>
          </Dropdown>
        </Layout.Header>
        <Layout.Content style={{ padding: 24 }}>
          <Routes>
            <Route path="/" element={<Navigate to="/routes" replace />} />
            <Route path="/routes" element={<RoutesPage />} />
            <Route path="/bot-decisions" element={<BotDecisionsPage />} />
            <Route path="/bot-alerts" element={<BotAlertsPage />} />
            <Route path="/waf-attacks" element={<WafAttackLogsPage />} />
          </Routes>
        </Layout.Content>
      </Layout>
    </Layout>
  );
}

export default function App() {
  const { token } = useAuth();
  return (
    <Routes>
      <Route path="/auth/callback" element={<AuthCallback />} />
      {token ? (
        <Route path="/*" element={<Shell />} />
      ) : (
        <Route path="/*" element={<LoginPage />} />
      )}
    </Routes>
  );
}
