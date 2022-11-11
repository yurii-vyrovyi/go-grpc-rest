

.PHONY: deps
deps:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run --allow-parallel-runners -v -c .golangci.yml


.PHONY: protogen_stub
protogen_stub:
	protoc -I ./api \
	  --go_out ./api --go_opt paths=source_relative \
	  --go-grpc_out ./api --go-grpc_opt paths=source_relative \
	  ./api/service.proto

.PHONY: protogen_gw
protogen_gw:
	protoc -I ./api \
	  --grpc-gateway_out ./api \
	  --grpc-gateway_opt logtostderr=true \
	  --grpc-gateway_opt paths=source_relative \
	  --grpc-gateway_opt grpc_api_configuration=./api/service-grpc.yaml \
	  ./api/service.proto

.PHONY: protogen_swag
protogen_swag:
	protoc -I ./api \
		--openapiv2_out ./api \
		--openapiv2_opt grpc_api_configuration=./api/service-grpc.yaml \
		./api/service.proto


.PHONY: protogen
protogen: protogen_stub protogen_gw protogen_swag


.PHONY: swag
swag:
	swag init --generalInfo ./api/service.pb.gw.go --parseVendor=true