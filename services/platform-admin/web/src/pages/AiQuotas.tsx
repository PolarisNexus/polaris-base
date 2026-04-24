import { useEffect, useState } from "react";
import { Typography, Alert, Space, Empty } from "antd";
import { aiGatewayApi, Quota } from "../api/aiGateway";

export default function AiQuotasPage() {
  const [rows, setRows] = useState<Quota[]>([]);
  const [note, setNote] = useState<string>();
  const [err, setErr] = useState<string>();

  useEffect(() => {
    aiGatewayApi
      .listQuotas()
      .then((r) => {
        setRows(r.items ?? []);
        setNote(r.phaseNote);
      })
      .catch((e: Error) => setErr(e.message));
  }, []);

  return (
    <Space direction="vertical" style={{ width: "100%" }} size="middle">
      <Typography.Title level={3}>AI 配额</Typography.Title>
      {err && <Alert type="error" message={err} showIcon />}
      {note && <Alert type="info" message={note} showIcon />}
      {rows.length === 0 && (
        <Empty description="暂无配额规则（ADR-0014 Phase II 将启用 ai-rate-limiting + CRUD UI）" />
      )}
    </Space>
  );
}
