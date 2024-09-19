# delete the templates code start
.PHONY: install
# installation of dependent plugins
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/srikrsna/protoc-gen-gotag@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@v1.8.12
	go install github.com/ofabry/go-callvis@latest
	go install golang.org/x/pkgsite/cmd/pkgsite@latest

.PHONY: proto
# generate *.go and template code by proto files, if you do not refer to the proto file, the default is all the proto files in the api directory. you can specify the proto file, multiple files are separated by commas, e.g. make proto FILES=api/user/v1/user.proto, only for ⓶ Microservices created based on sql, ⓷ Web services created based on protobuf, ⓸ Microservices created based on protobuf, ⓹ grpc gateway service created based on protobuf
proto:
	@bash scripts/protoc.sh $(FILES)
	go mod tidy
	@gofmt -s -w .
