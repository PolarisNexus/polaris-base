# api/proto/ — gRPC Protobuf 定义

统一维护平台所有 `.proto` 文件，各服务仓库引用此目录生成对应语言的桩代码。

## 目录约定

按服务域名建立子目录，每个子目录内按模块组织：

```
proto/
├── user/               ← 用户服务
│   └── v1/
│       └── user.proto
├── tenant/             ← 租户服务
│   └── v1/
│       └── tenant.proto
└── common/             ← 公共消息类型
    └── v1/
        └── types.proto
```

## 引用方式

- **Java (Maven)**：通过 `protobuf-maven-plugin` 引用本仓库 proto 目录生成代码
- **Python**：通过 `grpcio-tools` 或 `buf generate` 生成桩代码

## 命名规范

- 包名：`polaris.<domain>.v<N>`
- 文件名：小写下划线（snake_case）
- 消息/服务名：大驼峰（PascalCase）
