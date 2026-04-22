import { useCallback, useEffect, useState } from "react";
import {
  Alert,
  Button,
  DatePicker,
  Descriptions,
  Drawer,
  Input,
  Space,
  Table,
  Tag,
  Typography,
} from "antd";
import type { Dayjs } from "dayjs";
import { AttackLog, wafApi } from "../api/waf";

const { RangePicker } = DatePicker;

const ACTION_COLORS: Record<string, string> = {
  block: "red",
  reject: "orange",
  error: "volcano",
  log: "blue",
};

export default function WafAttackLogsPage() {
  const [rows, setRows] = useState<AttackLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string>();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [clientIp, setClientIp] = useState("");
  const [query, setQuery] = useState("");
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [detail, setDetail] = useState<AttackLog>();

  const refresh = useCallback(() => {
    setLoading(true);
    wafApi
      .queryAttackLogs({
        clientIp: clientIp || undefined,
        query: query || undefined,
        timeRange: range
          ? { start: range[0].toISOString(), end: range[1].toISOString() }
          : undefined,
        page: { page, pageSize },
        order: "SORT_ORDER_DESC",
      })
      .then((r) => {
        setRows(r.items ?? []);
        setTotal(Number(r.pageInfo?.total ?? 0));
      })
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  }, [clientIp, query, range, page, pageSize]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return (
    <Space direction="vertical" style={{ width: "100%" }} size="middle">
      <Typography.Title level={3}>
        WAF 攻击日志
        <Typography.Text type="secondary" style={{ fontSize: 14, marginLeft: 12 }}>
          来源：APISIX → Elasticsearch · MVP 口径 status ≥ 400
        </Typography.Text>
      </Typography.Title>

      <Space wrap>
        <RangePicker showTime onChange={(v) => setRange(v as [Dayjs, Dayjs] | null)} />
        <Input.Search
          placeholder="按 IP 精确匹配"
          style={{ width: 220 }}
          allowClear
          onSearch={setClientIp}
        />
        <Input.Search
          placeholder="按 URI / Host 模糊搜索"
          style={{ width: 320 }}
          allowClear
          onSearch={setQuery}
        />
        <Button onClick={refresh}>刷新</Button>
      </Space>

      {err && <Alert type="error" message={err} showIcon closable onClose={() => setErr(undefined)} />}

      <Table<AttackLog>
        rowKey={(r, i) => `${r.requestId || i}`}
        loading={loading}
        dataSource={rows}
        onRow={(r) => ({ onClick: () => setDetail(r), style: { cursor: "pointer" } })}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
        columns={[
          { title: "时间", dataIndex: "timestamp", width: 200 },
          {
            title: "动作",
            dataIndex: "action",
            width: 90,
            render: (v?: string) => <Tag color={ACTION_COLORS[v ?? "log"]}>{v ?? "-"}</Tag>,
          },
          { title: "Client IP", dataIndex: "clientIp", width: 140 },
          { title: "Method", dataIndex: "method", width: 80 },
          { title: "Host", dataIndex: "host", width: 180 },
          { title: "URI", dataIndex: "uri", ellipsis: true },
          { title: "Route", dataIndex: "routeId", width: 160 },
        ]}
      />

      <Drawer
        title={detail ? `请求 ${detail.requestId ?? ""}` : ""}
        open={!!detail}
        onClose={() => setDetail(undefined)}
        width={560}
      >
        {detail && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="时间">{detail.timestamp}</Descriptions.Item>
            <Descriptions.Item label="动作">
              <Tag color={ACTION_COLORS[detail.action ?? "log"]}>{detail.action}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="Client IP">{detail.clientIp}</Descriptions.Item>
            <Descriptions.Item label="Method">{detail.method}</Descriptions.Item>
            <Descriptions.Item label="Host">{detail.host}</Descriptions.Item>
            <Descriptions.Item label="URI">{detail.uri}</Descriptions.Item>
            <Descriptions.Item label="Route">{detail.routeId}</Descriptions.Item>
            <Descriptions.Item label="Rule ID">{detail.ruleId ?? "-"}</Descriptions.Item>
            <Descriptions.Item label="Rule 消息">{detail.ruleMessage || "-"}</Descriptions.Item>
            <Descriptions.Item label="Severity">{detail.severity || "-"}</Descriptions.Item>
            <Descriptions.Item label="命中片段">
              {detail.matchedData || (
                <Typography.Text type="secondary">
                  rule_id / severity / matched_data 需 Coraza 扩展响应头后才填充（ADR-0012 扩展项）
                </Typography.Text>
              )}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </Space>
  );
}
