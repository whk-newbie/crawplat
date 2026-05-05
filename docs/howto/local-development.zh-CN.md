# 本地开发指南

## 前置依赖

- Go 1.21+
- Node.js 18+
- Docker 与 Docker Compose
- Make

## 快速开始

### 1. 克隆仓库

```bash
git clone <repo-url> crawler-platform
cd crawler-platform
```

### 2. 启动开发环境

```bash
make dev-up
```

该命令会：
- 启动 PostgreSQL 并执行数据库迁移
- 以开发模式启动所有服务（Go 服务通过 `air` 热重载）
- 启动前端 Vite 开发服务器

### 3. 访问应用

- 前端：`http://localhost:3000`
- 网关 API：`http://localhost:8080`

### 4. 停止环境

```bash
make dev-down
```

## 项目结构

```
.
├── apps/                   # 应用服务
│   ├── gateway/            # API 网关
│   ├── iam-service/        # 认证服务
│   ├── project-service/    # 项目服务
│   ├── spider-service/     # 爬虫服务
│   ├── execution-service/  # 执行服务
│   ├── scheduler-service/  # 调度服务
│   ├── node-service/       # 节点服务
│   ├── datasource-service/ # 数据源服务
│   ├── monitor-service/    # 监控服务
│   └── web/                # 前端应用
├── deploy/                 # 部署配置
│   ├── docker-compose/     # Compose 文件
│   ├── migrations/         # 数据库迁移
│   └── scripts/            # 运维脚本
├── docs/                   # 文档
├── examples/               # 示例爬虫
│   └── spiders/
└── packages/               # 共享 Go 包
    └── go-common/
```

## 开发工作流

### 后端开发

修改 Go 服务代码后，`air` 会自动重新编译并重启对应服务。你修改以下路径的代码时会触发重载：

- `apps/*`（各服务）
- `packages/go-common`（共享包）

### 前端开发

前端使用 Vite 开发服务器，支持热模块替换（HMR）。修改 `apps/web/src` 下的文件后，浏览器会自动刷新。

### API 代理

前端开发服务器将 `/api/*` 请求代理到 Docker 网络中的 `gateway` 容器，无需手动配置 API 地址。

### 数据库迁移

如果修改了数据库表结构，需要添加迁移文件：

```bash
# 迁移文件位于
deploy/migrations/postgres/
```

然后重启开发环境：

```bash
make dev-down && make dev-up
```

## 环境变量

主要环境变量定义在 `deploy/env/.env.example`。各服务特定覆盖项：

| 变量 | 服务 | 默认值 | 说明 |
|----------|---------|---------|-------------|
| `DATABASE_DSN` | iam-service | *(空)* | 用户持久化的 PostgreSQL DSN；未设置时回退到内存存储（仅开发环境） |
| `IAM_ENABLE_SEED_ADMIN` | iam-service | `false` | 设为 `true` 时，在 users 表为空时自动创建 admin/admin123 |
| `IMAGE_REGISTRY_AUTH_MAP` | agent | *(空)* | 私有仓库凭据 JSON 映射，如 `{"ghcr.io":{"username":"u","password":"p"}}` |
| `AGENT_CAPABILITIES` | agent | `docker` | 心跳中上报的节点能力，逗号分隔 |
| `INTERNAL_API_TOKEN` | 全部 | `change-me` | 服务间内部认证 Token |
| `JWT_SECRET` | iam-service, gateway | `change-me` | JWT Token 签名与验证密钥 |

## 运行测试

### 全部测试

```bash
make test
```

### 仅后端测试

```bash
go test ./...
```

### 仅前端测试

```bash
npm --prefix apps/web test
```

## 构建示例爬虫镜像

```bash
docker build -t crawler/go-echo:latest examples/spiders/go-echo
docker build -t crawler/python-echo:latest examples/spiders/python-echo
```

## 完整构建与冒烟验证

如需验证发布形态，使用以下命令：

```bash
make migrate
make up
bash deploy/scripts/smoke-mvp.sh
```

详见 `docs/product/mvp-smoke-checklist.zh-CN.md`。
