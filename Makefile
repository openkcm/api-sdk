.PHONY: generate
generate: fetch-protos format go-gen

.PHONY: clean
clean:
	@find ./proto -type f -name '*.go' -exec rm {} +

.PHONY: fetch-protos
fetch-protos:
	@protofetch -o vendor-proto fetch

.PHONY: go-gen
go-gen: clean
	@go install ./cmd/protoc-gen-go-eventbuilders
	@go install ./cmd/protoc-gen-go-otlpconverters
	@find ./proto -type f -iname '*.proto' -exec \
		protoc -I./proto -I./vendor-proto \
		--go_out=./proto \
		--go_opt=paths=import \
		--go_opt=module=github.com/openkcm/api-sdk/proto \
		--go-grpc_out=./proto \
		--go-grpc_opt=paths=import \
		--go-grpc_opt=module=github.com/openkcm/api-sdk/proto \
		{} +
	@protoc -I./proto -I./vendor-proto \
		--go-eventbuilders_out=./proto \
		--go-eventbuilders_opt=paths=import \
		--go-eventbuilders_opt=module=github.com/openkcm/api-sdk/proto \
		./proto/kms/api/cmk/auditlog/events/v1/events.proto
	@protoc -I./proto -I./vendor-proto \
		--go-otlpconverters_out=./proto \
		--go-otlpconverters_opt=paths=import \
		--go-otlpconverters_opt=module=github.com/openkcm/api-sdk/proto \
		./proto/kms/api/cmk/auditlog/events/v1/events.proto
	@go mod tidy

.PHONY: install-tools
install-tools:
	brew install protobuf
	go install \
		google.golang.org/protobuf/cmd/protoc-gen-go@latest \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install ./cmd/protoc-gen-go-eventbuilders
	go install ./cmd/protoc-gen-go-otlpconverters
	brew install bufbuild/buf/buf
	npm install -g @coralogix/protofetch

.PHONY: test
test:
	go test -count=1 -race -coverprofile cover.out ./...

.PHONY: validate
validate: format lint breaking

.PHONY: lint
lint:
	@buf lint

.PHONY: breaking
breaking:
	@buf breaking --against https://github.com/openkcm/api-sdk.git#branch=main

.PHONY: format
format:
	@buf format -w

.PHONY: reuse-lint
reuse-lint:
	docker run --rm --volume $(PWD):/data fsfe/reuse lint
