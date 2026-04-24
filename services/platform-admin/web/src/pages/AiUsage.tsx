import { useEffect, useState } from "react";
import { Table, Typography, Alert, Tag, Space, Row, Col, Statistic, Select, Input, Button, Card } from "antd";
import { aiGatewayApi, QueryUsageReq, UsageRecord, UsageSummary } from "../api/aiGateway";

const PROVIDERS = ["", "openai", "claude", "deepseek", "qwen"];

export default function AiUsagePage() {
  const [items, setItems] = useState<UsageRecord[]>([]);
  const [summary, setSummary] = useState<UsageSummary | undefined>();
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string>();
  const [filter, setFilter] = useState<QueryUsageReq>({
    page: { page: 1, pageSize: 50 },
    order: "SORT_ORDER_DESC",
  });

  const load = (f: QueryUsageReq) => {
    setLoading(true);
    setErr(undefined);
    aiGatewayApi
      .queryUsage(f)
      .then((r) => {
        setItems(r.items ?? []);
        setSummary(r.summary);
      })
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load(filter);
  }, []);

  const apply = () => load(filter);

  return (
    <Space direction="vertical" style={{ width: "100%" }} size="middle">
      <Typography.Title level={3}>AI 用量</Typography.Title>

      {summary && (
        <Row gutter={16}>
          <Col span={6}>
            <Card><Statistic title="请求总数" value={summary.totalRequests ?? 0} /></Card>
          </Col>
          <Col span={6}>
            <Card><Statistic title="Prompt tokens" value={summary.totalPromptTokens ?? 0} /></Card>
          </Col>
          <Col span={6}>
            <Card><Statistic title="Completion tokens" value={summary.totalCompletionTokens ?? 0} /></Card>
          </Col>
          <Col span={6}>
            <Card>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>按 Provider</Typography.Text>
              <div style={{ marginTop: 8 }}>
                {Object.entries(summary.requestsByProvider ?? {}).map(([k, v]) => (
                  <Tag key={k} color="blue">{k}: {v}</Tag>
                ))}
              </div>
            </Card>
          </Col>
        </Row>
      )}

      <Space wrap>
        <Select
          style={{ width: 140 }}
          value={filter.provider ?? ""}
          onChange={(v) => setFilter({ ...filter, provider: v || undefined })}
          options={PROVIDERS.map((p) => ({ value: p, label: p || "全部 provider" }))}
        />
        <Input
          placeholder="按 user 过滤"
          allowClear
          style={{ width: 220 }}
          onChange={(e) => setFilter({ ...filter, user: e.target.value || undefined })}
        />
        <Input
          placeholder="按 model 过滤"
          allowClear
          style={{ width: 220 }}
          onChange={(e) => setFilter({ ...filter, model: e.target.value || undefined })}
        />
        <Button type="primary" onClick={apply}>查询</Button>
      </Space>

      {err && <Alert type="error" message={err} showIcon />}

      <Table<UsageRecord>
        rowKey={(r) => r.requestId ?? `${r.timestamp}-${r.user}-${r.model}`}
        loading={loading}
        dataSource={items}
        pagination={{ pageSize: 50 }}
        size="small"
        columns={[
          {
            title: "时间",
            dataIndex: "timestamp",
            width: 180,
            render: (t?: string) => (t ? new Date(t).toLocaleString("zh-CN") : "-"),
          },
          { title: "User", dataIndex: "user", width: 180, ellipsis: true },
          {
            title: "Provider",
            dataIndex: "provider",
            width: 100,
            render: (p?: string) => p ? <Tag color="blue">{p}</Tag> : "-",
          },
          { title: "Model", dataIndex: "model", width: 180, ellipsis: true },
          { title: "Prompt tok", dataIndex: "promptTokens", width: 100 },
          { title: "Completion tok", dataIndex: "completionTokens", width: 120 },
          { title: "Total tok", dataIndex: "totalTokens", width: 100 },
          {
            title: "Status",
            dataIndex: "statusCode",
            width: 80,
            render: (s?: number) => (
              <Tag color={s && s < 400 ? "green" : s && s < 500 ? "orange" : "red"}>{s ?? "-"}</Tag>
            ),
          },
          {
            title: "Latency",
            dataIndex: "latencyMs",
            width: 100,
            render: (v?: number) => (v ? `${Math.round(v)}ms` : "-"),
          },
        ]}
      />
    </Space>
  );
}
