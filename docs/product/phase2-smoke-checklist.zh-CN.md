# Phase 2 冒烟检查清单

## 准备镜像

启动服务栈前，先构建示例爬虫镜像：

```bash
docker build -t crawler/go-echo:latest examples/spiders/go-echo
docker build -t crawler/python-echo:latest examples/spiders/python-echo
```

## 启动服务栈

从仓库根目录执行：

```bash
make migrate
make up
```

`make migrate` 启动 PostgreSQL、等待就绪并执行 SQL 迁移。
`make up` 构建服务二进制文件、构建 Web 资源、重新执行迁移并启动 Compose 服务栈。

## 手动冒烟流程

1. 打开 `http://localhost:3000`。
2. 进入「项目」并创建一个项目。
3. 进入「爬虫」并为该项目注册一个 Docker 爬虫：
   - Go 示例镜像：`crawler/go-echo:latest`
   - Python 示例镜像：`crawler/python-echo:latest`
4. 进入「执行」并使用项目 ID、爬虫 ID、镜像和命令创建手动执行。
5. 确认 Agent 领取了该执行，且执行状态从 `pending` → `running` → `succeeded`。
6. 打开执行详情页，确认日志可见。

## 预期运行时连接关系

- PostgreSQL 存储项目、爬虫、数据源和执行元数据。
- Redis 存储节点存活状态和执行队列状态。
- MongoDB 存储执行日志。
- Agent 使用：
  - `NODE_SERVICE_URL=http://node-service:8084`
  - `EXECUTION_SERVICE_URL=http://execution-service:8085`
  - `INTERNAL_API_TOKEN` 访问执行 worker 路由
- Agent 容器现在包含 Docker CLI 并挂载 `/var/run/docker.sock`。

## 验证证据

以上流程于 2026-04-22 在当前环境中验证通过：

```text
Go 示例：
  project_id=e7ab72ff-45a6-4e43-9520-bbaa68b92af4
  spider_id=45d697ef-d579-4b9a-8c1f-95b53b8709aa
  execution_id=2b374bd7-7b14-4c6f-9026-2837b1ea4c2e
  status=succeeded
  logs=go spider started / go spider finished

Python 示例：
  project_id=bb33593a-d607-4d12-b001-8c98e79afa04
  spider_id=15fb0de2-ff4a-46dd-bc1a-0370851038c2
  execution_id=09adb472-aecc-481f-a2ff-0fa05b6612c3
  status=succeeded
  logs=python spider started / python spider finished
```

这表明当前的 Compose 服务栈、数据库迁移、Agent Poller、Docker 运行时路径和基于 MongoDB 的日志检索均已通过验证。
