# Docker Compose 开发工作流说明

这个项目现在同时保留两套 Compose 工作流：

- `deploy/docker-compose/docker-compose.mcp.yml`
  这是偏发布形态的运行方式，适合做完整构建和验证。
- `deploy/docker-compose/docker-compose.dev.yml`
  这是偏开发形态的运行方式，适合边改代码边调试。

## 适用场景

如果你要做日常开发，优先使用开发工作流。它解决的是两个问题：

- Go 服务改代码后自动重新编译和重启
- 前端页面改代码后通过 Vite 热更新立即生效

## 启动方式

在项目根目录执行：

```bash
make dev-up
```

这个命令会先启动开发栈里的 PostgreSQL 并执行迁移，然后拉起完整的开发环境。

## 停止方式

```bash
make dev-down
```

## 端口说明

- `http://localhost:8080`
  对外网关入口
- `http://localhost:3000`
  前端 Vite 开发服务器

浏览器访问 `http://localhost:3000` 即可。

## 热更新机制

### Go 服务

Go 服务运行在统一的开发镜像里，内部通过 `air` 监听源码目录。你修改这些路径下的代码时会触发自动重编译和重启：

- `apps/*`
- `packages/go-common`

### 前端

前端使用 Vite dev server。你修改 `apps/web/src` 下的文件后，页面会自动刷新或热替换。

## API 转发

前端开发服务器会把 `/api/*` 请求代理到 Docker 网络内的 `gateway` 容器，所以前端代码里不需要切换成手写的容器地址。

也就是说：

- 浏览器访问的是 `localhost:3000`
- Vite 在容器里把 `/api/*` 转发到 `http://gateway:8080`

## 开发建议

- 日常前后端联调时，优先使用 `make dev-up`
- 需要验证发布形态时，再使用 `make up`
- 如果修改了数据库表结构，先补迁移文件，再重新执行 `make dev-up`

## 当前限制

- 这套开发工作流优先解决“本地开发效率”，不是生产部署方案
- `agent` 仍然依赖宿主机的 Docker Socket
- 目前没有把这套开发流程接到远端 CI，只是本地容器化开发流
