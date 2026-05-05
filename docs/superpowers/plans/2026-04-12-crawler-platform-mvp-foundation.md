# Crawler Platform MVP Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first working vertical slice of the crawler platform: login, project management, spider registration, manual execution, agent heartbeat, execution logs, and datasource configuration preview.

**Architecture:** Use a monorepo with a Vue frontend, Go microservices, and a Go agent. Implement only the minimum service set required to prove the control-plane to execution-plane loop: `web`, `gateway`, `iam-service`, `project-service`, `spider-service`, `execution-service`, `node-service`, `datasource-service`, and `agent`. Use PostgreSQL for platform truth, Redis for coordination, and MongoDB for logs.

**Tech Stack:** Vue 3, TypeScript, Vite, Pinia, Vue Router, Go, Gin or Fiber, pgx, go-redis, mongo-go-driver, Docker Compose, Makefile.

---

## Progress Status

Last updated: 2026-04-12

Completed:

- [x] Task 1: Initialize Monorepo Skeleton
- [x] Task 2: Create Shared Go Foundation Package
- [x] Task 3: Implement IAM Service Login Slice
- [x] Task 4: Implement Project Service CRUD Slice
- [x] Task 5: Implement Spider Service CRUD Slice
- [x] Task 6: Implement Node Service and Agent Heartbeat Loop
- [x] Task 7: Implement Manual Execution and Log Ingest
- [x] Task 8: Implement Datasource Service MVP
- [x] Task 9: Build Gateway Service and Route Wiring
- [x] Task 10: Build Vue Web MVP Shell
- [x] Task 11: Wire Docker Compose and End-to-End Smoke Flow
- [x] Task 12: Documentation and Developer Onboarding

Verification notes:

- `make test` passes across the Go services, shared packages, and `apps/web`.
- `make up` is wired for a normal Docker host and the Compose stack is ready, but full runtime verification is blocked in this sandbox because Docker bridge networking cannot be created (`failed to add the host <=> sandbox pair interfaces: operation not supported`).

## Scope Split

The approved spec covers multiple independent subsystems. This plan only covers the MVP foundation needed for a testable first release. Follow-up plans should be written separately for:

- scheduled tasks and retry orchestration
- monitoring dashboards and aggregated metrics
- spider versioning hardening
- audit logging and richer RBAC
- datasource schema browsing enhancements

## File Structure

Planned repository structure for this MVP:

- Create: `apps/web/`
- Create: `apps/gateway/`
- Create: `apps/iam-service/`
- Create: `apps/project-service/`
- Create: `apps/spider-service/`
- Create: `apps/execution-service/`
- Create: `apps/node-service/`
- Create: `apps/datasource-service/`
- Create: `apps/agent/`
- Create: `packages/go-common/`
- Create: `deploy/docker-compose/docker-compose.mcp.yml`
- Create: `deploy/env/.env.example`
- Create: `Makefile`
- Create: `README.md`
- Test: `apps/*/internal/.../*_test.go`
- Test: `apps/web/src/**/__tests__/*`

### Service Responsibility Map

- `apps/gateway`: public HTTP entry, auth middleware passthrough, reverse proxy or API aggregation
- `apps/iam-service`: login, JWT issue, user lookup
- `apps/project-service`: project CRUD
- `apps/spider-service`: spider CRUD and runtime spec validation
- `apps/execution-service`: manual execution creation, status transitions, log ingest
- `apps/node-service`: node registration and heartbeat
- `apps/datasource-service`: datasource CRUD, connection test, preview query
- `apps/agent`: heartbeat loop, execution pull, local fake runner for MVP
- `packages/go-common`: config, logger, db clients, shared DTOs, auth helpers
- `apps/web`: login, project list, spider list, execution list, datasource pages

### MVP File Seeds

- Create: `packages/go-common/config/config.go`
- Create: `packages/go-common/httpx/response.go`
- Create: `packages/go-common/auth/jwt.go`
- Create: `packages/go-common/postgres/postgres.go`
- Create: `packages/go-common/redisx/redis.go`
- Create: `packages/go-common/mongox/mongo.go`
- Create: `apps/iam-service/cmd/server/main.go`
- Create: `apps/project-service/cmd/server/main.go`
- Create: `apps/spider-service/cmd/server/main.go`
- Create: `apps/execution-service/cmd/server/main.go`
- Create: `apps/node-service/cmd/server/main.go`
- Create: `apps/datasource-service/cmd/server/main.go`
- Create: `apps/agent/cmd/agent/main.go`
- Create: `apps/web/src/main.ts`

### Delivery Rule

Each task below should land independently and keep the repo runnable.

## Task 1: Initialize Monorepo Skeleton

**Files:**
- Create: `README.md`
- Create: `Makefile`
- Create: `go.work`
- Create: `deploy/env/.env.example`
- Create: `deploy/docker-compose/docker-compose.mcp.yml`

- [ ] **Step 1: Write the failing repository smoke checklist in README**

```md
# Crawler Platform

## Smoke Checks

- `make deps` should prepare frontend and backend dependencies
- `make test` should run Go and web tests
- `make up` should start PostgreSQL, Redis, MongoDB, gateway, iam-service, project-service, spider-service, execution-service, node-service, datasource-service, agent, and web
```

- [ ] **Step 2: Add the first failing orchestration target**

```make
.PHONY: test

test:
	go test ./...
```

- [ ] **Step 3: Run the initial test command to confirm the repo is still incomplete**

Run: `make test`
Expected: FAIL because no Go modules or packages exist yet

- [ ] **Step 4: Add the minimal workspace and environment skeleton**

```text
go 1.24.0
use (
	./packages/go-common
)
```

```env
POSTGRES_DSN=postgres://crawler:crawler@postgres:5432/crawler?sslmode=disable
REDIS_ADDR=redis:6379
MONGO_URI=mongodb://mongo:27017
JWT_SECRET=change-me
```

- [ ] **Step 5: Add the initial compose services skeleton**

```yaml
services:
  postgres:
    image: postgres:16
  redis:
    image: redis:7
  mongo:
    image: mongo:7
```

- [ ] **Step 6: Run the smoke test again**

Run: `make test`
Expected: PASS or `no packages to test` once the workspace is valid

- [ ] **Step 7: Commit**

```bash
git add README.md Makefile go.work deploy/env/.env.example deploy/docker-compose/docker-compose.mcp.yml
git commit -m "chore: initialize crawler platform workspace"
```

## Task 2: Create Shared Go Foundation Package

**Files:**
- Create: `packages/go-common/go.mod`
- Create: `packages/go-common/config/config.go`
- Create: `packages/go-common/httpx/response.go`
- Create: `packages/go-common/auth/jwt.go`
- Create: `packages/go-common/postgres/postgres.go`
- Create: `packages/go-common/redisx/redis.go`
- Create: `packages/go-common/mongox/mongo.go`
- Test: `packages/go-common/auth/jwt_test.go`

- [ ] **Step 1: Write the failing JWT helper test**

```go
package auth

import "testing"

func TestIssueAndParseToken(t *testing.T) {
	token, err := IssueToken("secret", "user-1")
	if err != nil {
		t.Fatalf("IssueToken returned error: %v", err)
	}

	claims, err := ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
}
```

- [ ] **Step 2: Run the focused test and confirm it fails**

Run: `go test ./packages/go-common/auth -run TestIssueAndParseToken -v`
Expected: FAIL because `IssueToken` and `ParseToken` do not exist yet

- [ ] **Step 3: Implement the minimal shared helpers**

```go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IssueToken(secret, subject string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}).SignedString([]byte(secret))
}

func ParseToken(secret, token string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	return claims, err
}
```

- [ ] **Step 4: Add minimal config and client wrappers**

```go
package config

type App struct {
	HTTPAddr    string
	PostgresDSN string
	RedisAddr   string
	MongoURI    string
	JWTSecret   string
}
```

- [ ] **Step 5: Run shared package tests**

Run: `go test ./packages/go-common/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add packages/go-common
git commit -m "feat: add shared go foundation package"
```

## Task 3: Implement IAM Service Login Slice

**Files:**
- Create: `apps/iam-service/go.mod`
- Create: `apps/iam-service/cmd/server/main.go`
- Create: `apps/iam-service/internal/api/router.go`
- Create: `apps/iam-service/internal/service/auth_service.go`
- Create: `apps/iam-service/internal/repo/user_repo.go`
- Create: `apps/iam-service/internal/model/user.go`
- Test: `apps/iam-service/internal/service/auth_service_test.go`

- [ ] **Step 1: Write the failing login service test**

```go
package service

import "testing"

func TestLoginReturnsTokenForSeedUser(t *testing.T) {
	svc := NewAuthService("secret")
	token, err := svc.Login("admin", "admin123")
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}
```

- [ ] **Step 2: Run the login service test and confirm it fails**

Run: `go test ./apps/iam-service/internal/service -run TestLoginReturnsTokenForSeedUser -v`
Expected: FAIL because `NewAuthService` does not exist yet

- [ ] **Step 3: Implement the minimal auth service with a seeded admin account**

```go
package service

import "crawler-platform/packages/go-common/auth"

type AuthService struct {
	secret string
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{secret: secret}
}

func (s *AuthService) Login(username, password string) (string, error) {
	if username != "admin" || password != "admin123" {
		return "", ErrInvalidCredentials
	}
	return auth.IssueToken(s.secret, "admin")
}
```

- [ ] **Step 4: Add the HTTP login route**

```go
router.POST("/api/v1/auth/login", func(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
})
```

- [ ] **Step 5: Run IAM tests**

Run: `go test ./apps/iam-service/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/iam-service
git commit -m "feat: add iam login service"
```

## Task 4: Implement Project Service CRUD Slice

**Files:**
- Create: `apps/project-service/go.mod`
- Create: `apps/project-service/cmd/server/main.go`
- Create: `apps/project-service/internal/api/router.go`
- Create: `apps/project-service/internal/model/project.go`
- Create: `apps/project-service/internal/service/project_service.go`
- Test: `apps/project-service/internal/service/project_service_test.go`

- [ ] **Step 1: Write the failing project creation test**

```go
package service

import "testing"

func TestCreateProjectAssignsID(t *testing.T) {
	svc := NewProjectService()
	project := svc.Create("core-crawlers", "Core Crawlers")
	if project.ID == "" {
		t.Fatal("expected generated id")
	}
}
```

- [ ] **Step 2: Run the project service test and confirm it fails**

Run: `go test ./apps/project-service/internal/service -run TestCreateProjectAssignsID -v`
Expected: FAIL because `NewProjectService` does not exist yet

- [ ] **Step 3: Implement in-memory project service**

```go
package service

import "github.com/google/uuid"

type Project struct {
	ID   string
	Code string
	Name string
}

type ProjectService struct{}

func NewProjectService() *ProjectService { return &ProjectService{} }

func (s *ProjectService) Create(code, name string) Project {
	return Project{ID: uuid.NewString(), Code: code, Name: name}
}
```

- [ ] **Step 4: Add create and list routes**

```go
router.POST("/api/v1/projects", createProjectHandler)
router.GET("/api/v1/projects", listProjectsHandler)
```

- [ ] **Step 5: Run project service tests**

Run: `go test ./apps/project-service/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/project-service
git commit -m "feat: add project service crud slice"
```

## Task 5: Implement Spider Service CRUD Slice

**Files:**
- Create: `apps/spider-service/go.mod`
- Create: `apps/spider-service/cmd/server/main.go`
- Create: `apps/spider-service/internal/api/router.go`
- Create: `apps/spider-service/internal/model/spider.go`
- Create: `apps/spider-service/internal/service/spider_service.go`
- Test: `apps/spider-service/internal/service/spider_service_test.go`

- [ ] **Step 1: Write the failing spider validation test**

```go
package service

import "testing"

func TestCreateSpiderRejectsUnknownLanguage(t *testing.T) {
	svc := NewSpiderService()
	_, err := svc.Create("p1", "bad", "ruby", "docker")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
```

- [ ] **Step 2: Run the spider service test and confirm it fails**

Run: `go test ./apps/spider-service/internal/service -run TestCreateSpiderRejectsUnknownLanguage -v`
Expected: FAIL because the service does not exist yet

- [ ] **Step 3: Implement minimal spider validation logic**

```go
func (s *SpiderService) Create(projectID, name, language, runtime string) (Spider, error) {
	if language != "go" && language != "python" {
		return Spider{}, ErrInvalidLanguage
	}
	if runtime != "docker" && runtime != "host" {
		return Spider{}, ErrInvalidRuntime
	}
	return Spider{ID: uuid.NewString(), ProjectID: projectID, Name: name, Language: language, Runtime: runtime}, nil
}
```

- [ ] **Step 4: Add create and list spider routes**

```go
router.POST("/api/v1/projects/:projectId/spiders", createSpiderHandler)
router.GET("/api/v1/projects/:projectId/spiders", listSpiderHandler)
```

- [ ] **Step 5: Run spider service tests**

Run: `go test ./apps/spider-service/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/spider-service
git commit -m "feat: add spider service crud slice"
```

## Task 6: Implement Node Service and Agent Heartbeat Loop

**Files:**
- Create: `apps/node-service/go.mod`
- Create: `apps/node-service/cmd/server/main.go`
- Create: `apps/node-service/internal/api/router.go`
- Create: `apps/node-service/internal/service/node_service.go`
- Create: `apps/agent/go.mod`
- Create: `apps/agent/cmd/agent/main.go`
- Create: `apps/agent/internal/heartbeat/heartbeat.go`
- Test: `apps/node-service/internal/service/node_service_test.go`

- [ ] **Step 1: Write the failing node heartbeat update test**

```go
package service

import "testing"

func TestHeartbeatMarksNodeOnline(t *testing.T) {
	svc := NewNodeService()
	node := svc.Heartbeat("node-a", []string{"docker", "python", "go"})
	if node.Status != "online" {
		t.Fatalf("expected online, got %s", node.Status)
	}
}
```

- [ ] **Step 2: Run the node service test and confirm it fails**

Run: `go test ./apps/node-service/internal/service -run TestHeartbeatMarksNodeOnline -v`
Expected: FAIL because `NewNodeService` does not exist yet

- [ ] **Step 3: Implement minimal node heartbeat service**

```go
func (s *NodeService) Heartbeat(name string, capabilities []string) Node {
	return Node{
		ID:           name,
		Name:         name,
		Status:       "online",
		Capabilities: capabilities,
		LastSeenAt:   time.Now(),
	}
}
```

- [ ] **Step 4: Implement agent heartbeat loop**

```go
func Run(ctx context.Context, baseURL, nodeName string) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			postHeartbeat(baseURL, nodeName)
		}
	}
}
```

- [ ] **Step 5: Run node and agent tests**

Run: `go test ./apps/node-service/... ./apps/agent/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/node-service apps/agent
git commit -m "feat: add node service and agent heartbeat loop"
```

## Task 7: Implement Manual Execution and Log Ingest

**Files:**
- Create: `apps/execution-service/go.mod`
- Create: `apps/execution-service/cmd/server/main.go`
- Create: `apps/execution-service/internal/api/router.go`
- Create: `apps/execution-service/internal/service/execution_service.go`
- Create: `apps/execution-service/internal/model/execution.go`
- Test: `apps/execution-service/internal/service/execution_service_test.go`

- [ ] **Step 1: Write the failing manual execution test**

```go
package service

import "testing"

func TestCreateManualExecutionStartsPending(t *testing.T) {
	svc := NewExecutionService()
	exec := svc.CreateManual("task-1", "spider-v1")
	if exec.Status != "pending" {
		t.Fatalf("expected pending, got %s", exec.Status)
	}
}
```

- [ ] **Step 2: Run the execution test and confirm it fails**

Run: `go test ./apps/execution-service/internal/service -run TestCreateManualExecutionStartsPending -v`
Expected: FAIL because `NewExecutionService` does not exist yet

- [ ] **Step 3: Implement execution state bootstrap**

```go
func (s *ExecutionService) CreateManual(taskID, spiderVersionID string) Execution {
	return Execution{
		ID:              uuid.NewString(),
		TaskID:          taskID,
		SpiderVersionID: spiderVersionID,
		Status:          "pending",
		TriggerSource:   "manual",
		CreatedAt:       time.Now(),
	}
}
```

- [ ] **Step 4: Add log ingest and execution detail routes**

```go
router.POST("/api/v1/executions", createExecutionHandler)
router.POST("/api/v1/executions/:id/logs", appendLogHandler)
router.GET("/api/v1/executions/:id", getExecutionHandler)
router.GET("/api/v1/executions/:id/logs", getLogsHandler)
```

- [ ] **Step 5: Run execution service tests**

Run: `go test ./apps/execution-service/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/execution-service
git commit -m "feat: add manual execution flow and log ingest"
```

## Task 8: Implement Datasource Service MVP

**Files:**
- Create: `apps/datasource-service/go.mod`
- Create: `apps/datasource-service/cmd/server/main.go`
- Create: `apps/datasource-service/internal/api/router.go`
- Create: `apps/datasource-service/internal/service/datasource_service.go`
- Create: `apps/datasource-service/internal/model/datasource.go`
- Test: `apps/datasource-service/internal/service/datasource_service_test.go`

- [ ] **Step 1: Write the failing datasource type validation test**

```go
package service

import "testing"

func TestCreateDatasourceRejectsUnknownType(t *testing.T) {
	svc := NewDatasourceService()
	_, err := svc.Create("project-1", "main", "mysql")
	if err == nil {
		t.Fatal("expected validation error")
	}
}
```

- [ ] **Step 2: Run the datasource test and confirm it fails**

Run: `go test ./apps/datasource-service/internal/service -run TestCreateDatasourceRejectsUnknownType -v`
Expected: FAIL because the service does not exist yet

- [ ] **Step 3: Implement minimal datasource validation**

```go
func (s *DatasourceService) Create(projectID, name, typ string) (Datasource, error) {
	switch typ {
	case "mongodb", "redis", "postgresql":
	default:
		return Datasource{}, ErrInvalidDatasourceType
	}
	return Datasource{ID: uuid.NewString(), ProjectID: projectID, Name: name, Type: typ, Readonly: true}, nil
}
```

- [ ] **Step 4: Add create, list, test, and preview routes**

```go
router.POST("/api/v1/datasources", createDatasourceHandler)
router.GET("/api/v1/datasources", listDatasourceHandler)
router.POST("/api/v1/datasources/:id/test", testDatasourceHandler)
router.POST("/api/v1/datasources/:id/preview", previewDatasourceHandler)
```

- [ ] **Step 5: Run datasource service tests**

Run: `go test ./apps/datasource-service/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/datasource-service
git commit -m "feat: add datasource service mvp"
```

## Task 9: Build Gateway Service and Route Wiring

**Files:**
- Create: `apps/gateway/go.mod`
- Create: `apps/gateway/cmd/server/main.go`
- Create: `apps/gateway/internal/api/router.go`
- Create: `apps/gateway/internal/proxy/proxy.go`
- Test: `apps/gateway/internal/proxy/proxy_test.go`

- [ ] **Step 1: Write the failing gateway target resolution test**

```go
package proxy

import "testing"

func TestResolveServiceURL(t *testing.T) {
	url := ResolveServiceURL("iam-service")
	if url == "" {
		t.Fatal("expected non-empty url")
	}
}
```

- [ ] **Step 2: Run the gateway test and confirm it fails**

Run: `go test ./apps/gateway/internal/proxy -run TestResolveServiceURL -v`
Expected: FAIL because `ResolveServiceURL` does not exist yet

- [ ] **Step 3: Implement minimal route forwarding map**

```go
func ResolveServiceURL(name string) string {
	switch name {
	case "iam-service":
		return "http://iam-service:8081"
	case "project-service":
		return "http://project-service:8082"
	default:
		return ""
	}
}
```

- [ ] **Step 4: Add gateway routes for auth, projects, spiders, executions, nodes, and datasources**

```go
router.Any("/api/v1/auth/*path", proxyTo("iam-service"))
router.Any("/api/v1/projects/*path", proxyTo("project-service"))
router.Any("/api/v1/executions/*path", proxyTo("execution-service"))
```

- [ ] **Step 5: Run gateway tests**

Run: `go test ./apps/gateway/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/gateway
git commit -m "feat: add gateway routing"
```

## Task 10: Build Vue Web MVP Shell

**Files:**
- Create: `apps/web/package.json`
- Create: `apps/web/vite.config.ts`
- Create: `apps/web/src/main.ts`
- Create: `apps/web/src/router/index.ts`
- Create: `apps/web/src/stores/auth.ts`
- Create: `apps/web/src/views/LoginView.vue`
- Create: `apps/web/src/views/ProjectsView.vue`
- Create: `apps/web/src/views/SpidersView.vue`
- Create: `apps/web/src/views/ExecutionsView.vue`
- Create: `apps/web/src/views/DatasourcesView.vue`
- Test: `apps/web/src/stores/__tests__/auth.spec.ts`

- [ ] **Step 1: Write the failing auth store test**

```ts
import { describe, expect, it } from 'vitest'
import { useAuthStore } from '../auth'

describe('auth store', () => {
  it('stores token after login success', async () => {
    const store = useAuthStore()
    store.setToken('token-1')
    expect(store.token).toBe('token-1')
  })
})
```

- [ ] **Step 2: Run the auth store test and confirm it fails**

Run: `cd apps/web && npm test -- auth.spec.ts`
Expected: FAIL because the store does not exist yet

- [ ] **Step 3: Implement the minimal Vue shell**

```ts
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({ token: '' }),
  actions: {
    setToken(token: string) {
      this.token = token
    },
  },
})
```

- [ ] **Step 4: Add router pages for login, projects, spiders, executions, and datasources**

```ts
[
  { path: '/login', component: LoginView },
  { path: '/projects', component: ProjectsView },
  { path: '/spiders', component: SpidersView },
  { path: '/executions', component: ExecutionsView },
  { path: '/datasources', component: DatasourcesView },
]
```

- [ ] **Step 5: Run web tests**

Run: `cd apps/web && npm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/web
git commit -m "feat: add web mvp shell"
```

## Task 11: Wire Docker Compose and End-to-End Smoke Flow

**Files:**
- Modify: `deploy/docker-compose/docker-compose.mcp.yml`
- Modify: `Makefile`
- Create: `deploy/scripts/smoke-mvp.sh`
- Create: `docs/product/mvp-smoke-checklist.md`

- [ ] **Step 1: Write the failing smoke script outline**

```bash
#!/usr/bin/env bash
set -euo pipefail

curl -f http://localhost:8080/health
curl -f http://localhost:8080/api/v1/auth/login
```

- [ ] **Step 2: Run the smoke script and confirm it fails before containers exist**

Run: `bash deploy/scripts/smoke-mvp.sh`
Expected: FAIL because the stack is not wired yet

- [ ] **Step 3: Add compose wiring for all MVP services**

```yaml
services:
  gateway:
    build: ../../apps/gateway
    ports: ["8080:8080"]
  iam-service:
    build: ../../apps/iam-service
  project-service:
    build: ../../apps/project-service
  spider-service:
    build: ../../apps/spider-service
  execution-service:
    build: ../../apps/execution-service
  node-service:
    build: ../../apps/node-service
  datasource-service:
    build: ../../apps/datasource-service
  agent:
    build: ../../apps/agent
```

- [ ] **Step 4: Add top-level make targets**

```make
up:
	docker compose -f deploy/docker-compose/docker-compose.mcp.yml up --build -d

down:
	docker compose -f deploy/docker-compose/docker-compose.mcp.yml down -v
```

- [ ] **Step 5: Run the smoke flow**

Run: `make up && bash deploy/scripts/smoke-mvp.sh && make down`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add deploy/docker-compose/docker-compose.mcp.yml Makefile deploy/scripts/smoke-mvp.sh docs/product/mvp-smoke-checklist.md
git commit -m "feat: wire mvp stack and smoke flow"
```

## Task 12: Documentation and Developer Onboarding

**Files:**
- Modify: `README.md`
- Create: `docs/architecture/mvp-overview.md`
- Create: `docs/api/mvp-service-map.md`

- [ ] **Step 1: Add the failing onboarding checklist**

```md
## Developer Checklist

- install Go and Node.js
- copy `deploy/env/.env.example`
- run `make up`
- run `make test`
```

- [ ] **Step 2: Verify the checklist against the working stack**

Run: `make test`
Expected: PASS

- [ ] **Step 3: Document the MVP architecture and service map**

```md
## MVP Service Map

- gateway -> public entry
- iam-service -> auth
- project-service -> project CRUD
- spider-service -> spider CRUD
- execution-service -> manual execution and logs
- node-service -> node heartbeat
- datasource-service -> datasource config and preview
- agent -> heartbeat and execution pull
```

- [ ] **Step 4: Run final repo verification**

Run: `make test`
Expected: PASS

Run: `make up`
Expected: all MVP containers healthy

- [ ] **Step 5: Commit**

```bash
git add README.md docs/architecture/mvp-overview.md docs/api/mvp-service-map.md
git commit -m "docs: add mvp onboarding and architecture notes"
```

## Self-Review

### Spec Coverage

Covered in this MVP plan:

- login and basic auth
- project boundary
- spider registration
- manual execution
- agent heartbeat
- execution logs
- datasource config and preview
- docker compose deployment

Deferred to follow-up plans:

- scheduled tasks
- retry policy
- aggregated monitoring dashboard
- richer RBAC
- audit log completeness
- spider version management hardening

### Placeholder Scan

This plan intentionally avoids `TODO`, `TBD`, and unresolved placeholders. Deferred items are explicitly called out as separate future plans rather than left ambiguous inside tasks.

### Type Consistency

Core names are kept consistent across tasks:

- `ProjectService`
- `SpiderService`
- `ExecutionService`
- `NodeService`
- `DatasourceService`
- `RuntimeSpec`
