SHELL=/bin/bash
NEW_GOPRIVATE=github.com/yurii-vyrovyi #,$(shell go env GOPRIVATE)
OSNAME=$(shell go env GOOS)


.PHONY: deps
deps:
	GOPRIVATE=$(NEW_GOPRIVATE)  go mod tidy

.PHONY: lint
lint:
	golangci-lint run --allow-parallel-runners -v -c .golangci.yml


