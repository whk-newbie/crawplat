# 常见问题排查

## 服务启动失败

### Go 服务编译错误

**症状**：`make dev-up` 后某服务反复重启或无法启动。

**排查**：

```bash
# 检查服务日志
docker compose -f deploy/docker-compose/docker-compose.dev.yml logs <service-name>

# 本地编译验证
go build ./apps/<service-name>/...
```

### 端口冲突

**症状**：启动时报端口已被占用。

**排查**：

```bash
# 检查端口占用
lsof -i :8080
lsof -i :3000
```

**解决**：停止占用端口的进程，或修改 `.env` 中的端口配置。

## 数据库问题

### 迁移失败

**症状**：服务启动时报数据库表不存在或字段缺失。

**解决**：

```bash
# 重新执行迁移
make dev-down
make migrate
make dev-up
```

### 连接被拒绝

**症状**：服务日志中出现 `connection refused`。

**排查**：

```bash
# 确认 PostgreSQL 容器是否运行
docker ps | grep postgres

# 检查数据库就绪状态
docker compose -f deploy/docker-compose/docker-compose.dev.yml exec postgres pg_isready
```

## 前端问题

### 页面白屏 / 无法加载

**排查**：

1. 确认 Vite 开发服务器已启动：`docker ps | grep web`
2. 检查浏览器控制台是否有网络错误
3. 确认网关可访问：`curl http://localhost:8080/api/v1/projects`

### 语言切换无效

**排查**：

1. 检查浏览器 localStorage 是否被清除
2. 确认 `localeStore` 中语言状态是否正常
3. 验证 `messages.ts` 中相关 key 是否存在

## Agent 问题

### Agent 无法领取任务

**排查**：

```bash
# 检查 Agent 日志
docker compose -f deploy/docker-compose/docker-compose.dev.yml logs agent

# 验证 Redis 队列
docker compose -f deploy/docker-compose/docker-compose.dev.yml exec redis redis-cli PING
```

### Docker 镜像拉取失败

**排查**：

1. 确认镜像是否已构建：`docker images | grep crawler`
2. 检查 Agent 容器是否可访问 Docker Socket
3. 对于私有仓库，确认 `IMAGE_REGISTRY_AUTH_MAP` 配置是否正确

## 测试问题

### 测试超时

**症状**：`make test` 中部分测试超时。

**解决**：

```bash
# 单独运行超时的测试包，增加超时时间
go test -timeout 60s ./apps/<service-name>/...
```

### 前端测试失败

**排查**：

```bash
# 在容器内运行前端测试
npm --prefix apps/web test
```
