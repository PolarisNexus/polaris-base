import { useEffect, useState } from "react";
import { Alert, Spin } from "antd";
import { handleCallback } from "../auth/oidc";

export default function AuthCallback() {
  const [err, setErr] = useState<string | null>(null);
  useEffect(() => {
    handleCallback()
      .then((returnTo) => {
        // hard redirect 以清掉地址栏上的 code/state 并重走 AuthProvider 读 token
        window.location.replace(returnTo || "/");
      })
      .catch((e: Error) => setErr(e.message));
  }, []);
  if (err) return <Alert type="error" message="登录失败" description={err} showIcon />;
  return (
    <div style={{ display: "flex", alignItems: "center", justifyContent: "center", padding: 48 }}>
      <Spin tip="正在登录..." />
    </div>
  );
}
