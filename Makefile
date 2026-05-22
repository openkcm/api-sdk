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
	@find ./proto -type f -iname '*.proto' -exec \
		protoc -I./proto -I./vendor-proto \
		--go_out=./proto \
		--go_opt=paths=import \
		--go_opt=module=github.com/openkcm/api-sdk/proto \
		--go-grpc_out=./proto \
		--go-grpc_opt=paths=import \
		--go-grpc_opt=module=github.com/openkcm/api-sdk/proto \
		{} +
	@go mod tidy

.PHONY: install-tools
install-tools:
	brew install protobuf
	go install \
		google.golang.org/protobuf/cmd/protoc-gen-go@latest \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	brew install bufbuild/buf/buf
	npm install -g @coralogix/protofetch

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
