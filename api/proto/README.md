# api/proto/ — gRPC Protobuf 定义

统一维护平台所有 `.proto` 文件，各服务仓库引用此目录生成对应语言的桩代码。

## 目录结构

```
proto/
└── polaris/
    ├── common/v1/types.proto              公共消息类型
    └── platform_admin/v1/                  platform-admin BFF 契约
        ├── gateway.proto
        ├── waf.proto
        └── bot.proto
```

## 工具链

- 契约管理：`buf`（`buf.yaml`、`buf.gen.yaml`）
- 生成产物：`api/gen/go/`（Go + Connect 桩），作为独立 Go module 由 `go.work` 聚合

## 命名规范

- 包名：`polaris.<domain>.v<N>`（如 `polaris.platform_admin.v1`）
- 文件名：小写下划线（snake_case）
- 消息/服务名：大驼峰（PascalCase）

## 约束

- 契约先行：先在此处定义接口，再到各服务实现
- 向后兼容：字段只增不删，废弃字段用 `reserved` 标记
- 规范校验：`buf lint` STANDARD；`buf breaking` FILE 级别
