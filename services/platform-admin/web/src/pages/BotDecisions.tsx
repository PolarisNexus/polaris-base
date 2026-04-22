import { useCallback, useEffect, useState } from "react";
import {
  Alert,
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import { botApi, Decision } from "../api/bot";

const SCOPES = [
  { label: "IP", value: "ip" },
  { label: "Range (CIDR)", value: "range" },
  { label: "Country", value: "country" },
  { label: "AS", value: "as" },
];

const TYPES = [
  { label: "Ban", value: "ban" },
  { label: "Captcha", value: "captcha" },
  { label: "Throttle", value: "throttle" },
];

interface CreateForm {
  scope: string;
  value: string;
  type: string;
  hours: number;
  reason: string;
}

export default function BotDecisionsPage() {
  const [rows, setRows] = useState<Decision[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string>();
  const [filter, setFilter] = useState<{ scope?: string; value?: string }>({});
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm<CreateForm>();

  const refresh = useCallback(() => {
    setLoading(true);
    botApi
      .listDecisions(filter)
      .then((r) => setRows(r.items ?? []))
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  }, [filter]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  async function submitCreate() {
    const v = await form.validateFields();
    try {
      await botApi.createDecision({
        scope: v.scope,
        value: v.value,
        type: v.type,
        duration: `${v.hours * 3600}s`,
        reason: v.reason,
      });
      message.success("封禁已下发");
      setOpen(false);
      form.resetFields();
      refresh();
    } catch (e) {
      message.error((e as Error).message);
    }
  }

  async function unban(r: Decision) {
    try {
      const resp = await botApi.deleteDecision({ id: r.id });
      message.success(`已解封 ${resp.deletedCount ?? 1} 条`);
      refresh();
    } catch (e) {
      message.error((e as Error).message);
    }
  }

  return (
    <Space direction="vertical" style={{ width: "100%" }} size="middle">
      <Typography.Title level={3}>Bot 决策（CrowdSec 当前生效封禁）</Typography.Title>

      <Space>
        <Select
          placeholder="作用域"
          style={{ width: 140 }}
          allowClear
          options={SCOPES}
          onChange={(scope) => setFilter((f) => ({ ...f, scope }))}
        />
        <Input.Search
          placeholder="按 value 过滤（如 IP）"
          style={{ width: 280 }}
          allowClear
          onSearch={(value) => setFilter((f) => ({ ...f, value }))}
        />
        <Button type="primary" onClick={() => setOpen(true)}>
          手动封禁
        </Button>
        <Button onClick={refresh}>刷新</Button>
      </Space>

      {err && <Alert type="error" message={err} showIcon closable onClose={() => setErr(undefined)} />}

      <Table<Decision>
        rowKey="id"
        loading={loading}
        dataSource={rows}
        pagination={{ pageSize: 20 }}
        columns={[
          { title: "ID", dataIndex: "id", width: 100 },
          {
            title: "来源",
            dataIndex: "origin",
            width: 140,
            render: (v: string) => <Tag>{v || "-"}</Tag>,
          },
          { title: "类型", dataIndex: "type", width: 90 },
          { title: "Scope", dataIndex: "scope", width: 90 },
          { title: "Value", dataIndex: "value" },
          { title: "Scenario", dataIndex: "scenario" },
          {
            title: "剩余时长",
            dataIndex: "duration",
            width: 120,
            render: (v?: string) => formatDuration(v),
          },
          {
            title: "操作",
            width: 120,
            render: (_, r) => (
              <Popconfirm title={`解封 ${r.scope}:${r.value}？`} onConfirm={() => unban(r)}>
                <Button danger size="small">
                  解封
                </Button>
              </Popconfirm>
            ),
          },
        ]}
      />

      <Modal
        title="手动封禁"
        open={open}
        onCancel={() => setOpen(false)}
        onOk={submitCreate}
        okText="下发"
        destroyOnClose
      >
        <Form<CreateForm>
          form={form}
          layout="vertical"
          initialValues={{ scope: "ip", type: "ban", hours: 4 }}
        >
          <Form.Item name="scope" label="作用域" rules={[{ required: true }]}>
            <Select options={SCOPES} />
          </Form.Item>
          <Form.Item
            name="value"
            label="Value"
            rules={[{ required: true, message: "请输入 IP / CIDR / 国家代码" }]}
          >
            <Input placeholder="例：1.2.3.4 或 10.0.0.0/24 或 CN" />
          </Form.Item>
          <Form.Item name="type" label="动作" rules={[{ required: true }]}>
            <Select options={TYPES} />
          </Form.Item>
          <Form.Item name="hours" label="时长（小时）" rules={[{ required: true }]}>
            <InputNumber min={1} max={24 * 30} style={{ width: "100%" }} />
          </Form.Item>
          <Form.Item name="reason" label="原因（写入审计）" rules={[{ required: true }]}>
            <Input.TextArea rows={2} />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}

function formatDuration(v?: string): string {
  if (!v) return "-";
  // Connect JSON 以 "14400s" 形式传。
  const m = v.match(/^(\d+(?:\.\d+)?)s$/);
  if (!m) return v;
  const sec = Number(m[1]);
  const h = Math.floor(sec / 3600);
  const m2 = Math.floor((sec % 3600) / 60);
  return h > 0 ? `${h}h${m2}m` : `${m2}m`;
}
