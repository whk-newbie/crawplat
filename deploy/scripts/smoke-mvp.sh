#!/usr/bin/env bash

set -euo pipefail

base_url="${BASE_URL:-http://localhost:8080}"

wait_for_gateway() {
  local attempts=30
  local sleep_seconds=2

  for ((i = 1; i <= attempts; i++)); do
    if curl --silent --show-error --fail "${base_url}/api/v1/projects" >/tmp/mvp-projects-response.json 2>/tmp/mvp-projects-error.log; then
      return 0
    fi
    sleep "${sleep_seconds}"
  done

  echo "Gateway did not become ready at ${base_url}/api/v1/projects" >&2
  if [[ -s /tmp/mvp-projects-error.log ]]; then
    cat /tmp/mvp-projects-error.log >&2
  fi
  return 1
}

assert_projects_list() {
  local body
  body="$(cat /tmp/mvp-projects-response.json)"
  if [[ "${body}" != "[]" ]]; then
    echo "Expected GET /api/v1/projects to return [] from the fresh MVP stack, got: ${body}" >&2
    return 1
  fi
  echo "Verified GET /api/v1/projects through gateway"
}

assert_seed_login() {
  local response
  response="$(
    curl --silent --show-error --fail \
      -H 'Content-Type: application/json' \
      -d '{"username":"admin","password":"admin123"}' \
      "${base_url}/api/v1/auth/login"
  )"

  if [[ "${response}" != *'"token"'* ]]; then
    echo "Expected seeded admin login to return a token, got: ${response}" >&2
    return 1
  fi

  echo "Verified POST /api/v1/auth/login through gateway"
}

wait_for_gateway
assert_projects_list
assert_seed_login

echo "MVP smoke checks passed"
