import { useEffect, useState } from "react";
import { Table, Typography, Alert, Tag, Space } from "antd";
import { aiGatewayApi, Provider } from "../api/aiGateway";

export default function AiProvidersPage() {
  const [rows, setRows] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string>();

  useEffect(() => {
    setLoading(true);
    aiGatewayApi
      .listProviders()
      .then((r) => setRows(r.items ?? []))
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  }, []);

  const statusColor = (s: string) =>
    ({ active: "green", dev: "gold", disabled: "red" }[s] ?? "default");

  return (
    <Space direction="vertical" style={{ width: "100%" }} size="middle">
      <Typography.Title level={3}>AI Providers（只读 · 源真相为 Git）</Typography.Title>
      <Typography.Paragraph type="secondary">
        路由映射来自 <code>components/apisix/routes/20-ai-gateway.yaml</code>。编辑请走 Git PR（ADR-0014）。
      </Typography.Paragraph>
      {err && <Alert type="error" message={err} showIcon />}
      <Table<Provider>
        rowKey="id"
        loading={loading}
        dataSource={rows}
        pagination={false}
        columns={[
          { title: "ID", dataIndex: "id", width: 120 },
          { title: "展示名", dataIndex: "displayName", width: 160 },
          {
            title: "APISIX provider",
            dataIndex: "apisixProvider",
            width: 160,
            render: (p: string) => <Tag color="blue">{p}</Tag>,
          },
          {
            title: "客户端 baseURL",
            dataIndex: "baseUrl",
            render: (u: string) => <code>{u}</code>,
          },
          {
            title: "支持路径",
            dataIndex: "supportedPaths",
            render: (paths: string[]) => (paths ?? []).map((p) => <Tag key={p}>{p}</Tag>),
          },
          {
            title: "状态",
            dataIndex: "status",
            width: 100,
            render: (s: string) => <Tag color={statusColor(s)}>{s}</Tag>,
          },
          {
            title: "上游 endpoint",
            dataIndex: "upstreamEndpoint",
            render: (u: string) => <code>{u}</code>,
          },
        ]}
      />
    </Space>
  );
}
