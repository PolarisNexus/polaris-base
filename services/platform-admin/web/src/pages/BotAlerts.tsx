import { useCallback, useEffect, useState } from "react";
import { Alert as AlertBox, Button, Descriptions, Drawer, Input, Space, Table, Tag, Typography } from "antd";
import { botApi, Alert } from "../api/bot";

export default function BotAlertsPage() {
  const [rows, setRows] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string>();
  const [filter, setFilter] = useState<{ sourceIp?: string; scenario?: string }>({});
  const [detail, setDetail] = useState<Alert>();

  const refresh = useCallback(() => {
    setLoading(true);
    botApi
      .listAlerts({ ...filter, page: { pageSize: 100 } })
      .then((r) => setRows(r.items ?? []))
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  }, [filter]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  async function openDetail(id: string) {
    try {
      const r = await botApi.getAlert(id);
      setDetail(r.alert);
    } catch (e) {
      setErr((e as Error).message);
    }
  }

  return (
    <Space direction="vertical" style={{ width: "100%" }} size="middle">
      <Typography.Title level={3}>Bot 告警（触发决策的原始事件）</Typography.Title>

      <Space>
        <Input.Search
          placeholder="按 IP 过滤"
          style={{ width: 260 }}
          allowClear
          onSearch={(sourceIp) => setFilter((f) => ({ ...f, sourceIp }))}
        />
        <Input.Search
          placeholder="按 scenario 过滤"
          style={{ width: 320 }}
          allowClear
          onSearch={(scenario) => setFilter((f) => ({ ...f, scenario }))}
        />
        <Button onClick={refresh}>刷新</Button>
      </Space>

      {err && <AlertBox type="error" message={err} showIcon closable onClose={() => setErr(undefined)} />}

      <Table<Alert>
        rowKey="id"
        loading={loading}
        dataSource={rows}
        pagination={{ pageSize: 20 }}
        onRow={(r) => ({ onClick: () => openDetail(r.id), style: { cursor: "pointer" } })}
        columns={[
          { title: "ID", dataIndex: "id", width: 100 },
          { title: "触发 IP", dataIndex: "sourceIp", width: 160 },
          { title: "Scenario", dataIndex: "scenario" },
          {
            title: "事件数",
            dataIndex: "eventsCount",
            width: 100,
            render: (v?: number) => <Tag color={v && v > 10 ? "red" : "blue"}>{v ?? 0}</Tag>,
          },
          { title: "开始时间", dataIndex: "startedAt", width: 200 },
          { title: "结束时间", dataIndex: "stoppedAt", width: 200 },
        ]}
      />

      <Drawer
        title={detail ? `告警 #${detail.id}` : ""}
        open={!!detail}
        onClose={() => setDetail(undefined)}
        width={520}
      >
        {detail && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="Scenario">{detail.scenario}</Descriptions.Item>
            <Descriptions.Item label="触发 IP">{detail.sourceIp}</Descriptions.Item>
            <Descriptions.Item label="Scope">{detail.sourceScope}</Descriptions.Item>
            <Descriptions.Item label="事件数">{detail.eventsCount}</Descriptions.Item>
            <Descriptions.Item label="开始">{detail.startedAt}</Descriptions.Item>
            <Descriptions.Item label="结束">{detail.stoppedAt}</Descriptions.Item>
            <Descriptions.Item label="关联决策">
              {(detail.decisionIds ?? []).map((id) => (
                <Tag key={id}>{id}</Tag>
              ))}
            </Descriptions.Item>
            <Descriptions.Item label="消息">{detail.message}</Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </Space>
  );
}
