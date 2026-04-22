import { Button, Card } from "antd";
import { useAuth } from "../auth/AuthContext";

export default function LoginPage() {
  const { login } = useAuth();
  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "#f0f2f5",
      }}
    >
      <Card style={{ width: 360, textAlign: "center" }}>
        <h2 style={{ marginBottom: 8 }}>Polaris Platform Admin</h2>
        <p style={{ color: "#888", marginBottom: 24 }}>通过 Authentik 单点登录</p>
        <Button type="primary" size="large" onClick={login} block>
          使用 Authentik 登录
        </Button>
      </Card>
    </div>
  );
}
