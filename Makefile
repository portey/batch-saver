export GO111MODULE=on
export GOSUMDB=off

.PHONE: build
build: dep
	go build -mod=vendor -o ./bin/${BIN_NAME} -a .

.PHONE: proto
proto:
	docker run --rm -v `pwd`:/defs  namely/protoc-all -i protos -f service.proto -l go -o gen/api

.PHONY: dep
dep:
	go mod tidy
	go mod download
	go mod vendor

.PHONY: test
test: dep
	go test -race -count=1 -short ./...

.PHONY: lint
lint: dep
	golangci-lint run -c .golangci.yml