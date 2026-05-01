module crawler-platform/apps/iam-service

go 1.25.0

require (
	crawler-platform/packages/go-common v0.0.0
	github.com/gin-gonic/gin v1.10.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
)

replace crawler-platform/packages/go-common => ../../packages/go-common
