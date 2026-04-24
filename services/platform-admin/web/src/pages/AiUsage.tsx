import { useEffect, useState } from "react";
import { Table, Typography, Alert, Tag, Space, Row, Col, Statistic, Select, Input, Button, Card, message } from "antd";
import { aiGatewayApi, QueryUsageReq, UsageRecord, UsageSummary } from "../api/aiGateway";
import { getToken } from "../auth/oidc";

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

  // 用当前登录会话的 JWT，直接打 APISIX 的 AI Gateway 一次 chat；
  // 回来后等 10s ES flush 再刷新 Usage 表。
  const sendTest = async (provider: string) => {
    const token = getToken();
    if (!token) {
      message.error("未登录或 token 已过期");
      return;
    }
    const modelByProvider: Record<string, string> = {
      openai: "gpt-4o-mini",
      claude: "claude-3-5-sonnet",
      deepseek: "deepseek-chat",
      qwen: "qwen-max",
    };
    try {
      // 走 Vite 代理 /ai/v1/* → APISIX :9080，避免浏览器 CORS
      const resp = await fetch(`/ai/v1/${provider}/chat/completions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          model: modelByProvider[provider],
          messages: [{ role: "user", content: `AI Gateway 冒烟测试 @ ${new Date().toLocaleTimeString("zh-CN")}` }],
        }),
      });
      if (!resp.ok) {
        message.error(`${provider} 返回 ${resp.status}`);
        return;
      }
      const data = await resp.json();
      const usage = data?.usage ?? {};
      message.success(
        `${provider} OK · model=${data?.model} · tokens=${usage.prompt_tokens}+${usage.completion_tokens}=${usage.total_tokens}`,
      );
      setTimeout(() => load(filter), 8000);
    } catch (e) {
      message.error(`${provider} 请求失败：${(e as Error).message}`);
    }
  };

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

      <Card size="small" title="AI Gateway 冒烟测试（用当前会话 JWT 发一次 chat）">
        <Space wrap>
          {PROVIDERS.filter((p) => p).map((p) => (
            <Button key={p} onClick={() => sendTest(p)}>
              测试 {p}
            </Button>
          ))}
          <Typography.Text type="secondary">
            点击后 ~8s 自动刷新下方表格
          </Typography.Text>
        </Space>
      </Card>

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
