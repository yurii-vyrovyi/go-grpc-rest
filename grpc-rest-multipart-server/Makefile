

.PHONY: deps
deps:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run --allow-parallel-runners -v -c .golangci.yml

.PHONY: protogen
protogen:
	protoc -I ./api \
	  --go_out ./api --go_opt paths=source_relative \
	  --go-grpc_out ./api --go-grpc_opt paths=source_relative \
	  ./api/grpc-rest-multipart-server.proto


