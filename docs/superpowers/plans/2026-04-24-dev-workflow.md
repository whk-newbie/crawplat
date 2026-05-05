# Dev Workflow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** add a Docker Compose based hot-reload development workflow without disturbing the existing production-shaped Compose flow.

**Architecture:** keep the current production Compose and add a dedicated dev Compose file, a reusable Go dev image, and Vite dev-server proxying. Public-facing docs stay in the main docs tree, while this plan remains local-only.

**Tech Stack:** Docker Compose, Go, Air, Vite, Node.js

---

### Task 1: Create Dev Compose Workflow

**Files:**
- Create: `deploy/docker-compose/docker-compose.dev.yml`
- Create: `deploy/docker/dev/go-dev.Dockerfile`
- Create: `deploy/docker/dev/run-go-service.sh`
- Modify: `Makefile`
- Modify: `deploy/scripts/migrate-postgres.sh`

### Task 2: Wire Frontend Dev Proxy

**Files:**
- Modify: `apps/web/vite.config.ts`

### Task 3: Document the Workflow Publicly

**Files:**
- Modify: `README.md`

### Task 4: 编写中文文档

**Files:**
- Create: `docs/product/docker-compose-dev-workflow.zh-CN.md`
- Modify: `README.md`
