module crawler-platform/apps/iam-service

go 1.24.0

require (
	crawler-platform/packages/go-common v0.0.0
	github.com/gin-gonic/gin v1.10.1
)

replace crawler-platform/packages/go-common => ../../packages/go-common
