#!/usr/bin/env sh
set -eu

: "${SERVICE_BINARY:?SERVICE_BINARY is required}"
: "${SERVICE_BUILD_TARGET:?SERVICE_BUILD_TARGET is required}"

cat >/tmp/air.toml <<EOF
root = "/workspace"
tmp_dir = "/tmp/air-${SERVICE_BINARY}"

[build]
cmd = "go build -o /tmp/${SERVICE_BINARY} ${SERVICE_BUILD_TARGET}"
bin = "/tmp/${SERVICE_BINARY}"
delay = 500
exclude_dir = ["apps/web/node_modules", "apps/web/dist", ".git", ".worktrees", ".docker-bin"]
include_ext = ["go"]
stop_on_error = true

[log]
time = true
EOF

exec air -c /tmp/air.toml
