.PHONY: test

NESTED_GO_MODULES := $(shell find packages apps -name go.mod -exec dirname {} \; | sort)

test:
	go test ./...
	@set -e; \
	for mod in $(NESTED_GO_MODULES); do \
		( cd "$$mod" && go test ./... ); \
	done
