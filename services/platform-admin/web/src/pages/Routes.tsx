import { useEffect, useState } from "react";
import { Table, Typography, Alert, Tag, Input, Space } from "antd";
import { listRoutes, Route } from "../api/gateway";

export default function RoutesPage() {
  const [rows, setRows] = useState<Route[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string>();
  const [query, setQuery] = useState("");

  useEffect(() => {
    setLoading(true);
    listRoutes(query)
      .then((r) => setRows(r.items ?? []))
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  }, [query]);

  return (
    <Space direction="vertical" style={{ width: "100%" }} size="middle">
      <Typography.Title level={3}>网关路由（只读 · 源真相为 Git）</Typography.Title>
      <Input.Search placeholder="按 URI / ID / 描述 过滤" allowClear onSearch={setQuery} />
      {err && <Alert type="error" message={err} showIcon />}
      <Table<Route>
        rowKey="id"
        loading={loading}
        dataSource={rows}
        pagination={{ pageSize: 20 }}
        columns={[
          { title: "ID", dataIndex: "id", width: 200 },
          { title: "URI", dataIndex: "uri" },
          {
            title: "Methods",
            dataIndex: "methods",
            render: (m?: string[]) => (m ?? []).map((x) => <Tag key={x}>{x}</Tag>),
            width: 200,
          },
          { title: "Upstream", dataIndex: "upstreamId", width: 180 },
          { title: "描述", dataIndex: "desc" },
        ]}
      />
    </Space>
  );
}
