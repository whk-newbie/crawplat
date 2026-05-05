# MVP 冒烟检查清单

## 启动服务栈

从仓库根目录启动完整 MVP 服务栈：

```bash
make up
```

`make up` 首先将 Linux 二进制文件编译到 `.docker-bin/` 目录，然后基于这些产物构建 Compose 镜像。
同时执行 `npm --prefix apps/web run build`，使 Web 容器提供最新构建的静态资源。
Compose 服务栈使用标准服务网络：网关暴露在 `http://localhost:8080`，Web Shell 暴露在 `http://localhost:3000`。

完成后停止并清理服务栈：

```bash
make down
```

## 执行冒烟检查

服务栈运行后执行：

```bash
bash deploy/scripts/smoke-mvp.sh
```

脚本会等待网关在 `http://localhost:8080` 上开始提供服务，然后验证：

- `GET /api/v1/projects` 通过网关成功返回，从新启动的 MVP 服务栈返回空项目列表
- `POST /api/v1/auth/login` 通过网关成功返回，使用预置管理员账号 `admin` / `admin123` 登录
- `GET /api/v1/datasources` 通过网关成功返回，从新启动的 MVP 服务栈返回空数据源列表
- Web 容器的 `GET /` 成功返回 `Crawler Platform` HTML 壳

## 说明

- Compose 服务栈启用 `IAM_ENABLE_SEED_ADMIN=true` 并提供开发用 `JWT_SECRET`，使预置账号可直接登录，无需额外配置。
- Agent 使用 `NODE_SERVICE_URL=http://node-service:8084` 和 `NODE_NAME=mvp-node`，可通过标准 Compose DNS 解析节点服务。
