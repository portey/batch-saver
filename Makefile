export GO111MODULE=on
export GOSUMDB=off

IMAGE_TAG := $(shell git rev-parse HEAD)
DOCKER_COMPOSE = docker-compose -f docker-compose.yml

.PHONE: build
build: dep
	go build -mod=vendor -o ./bin/svc -a .

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

.PHONY: dockerise
dockerise:
	docker build -t "batch-saver:${IMAGE_TAG}" .

.PHONY: docker-up
docker-up: dockerise
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	$(DOCKER_COMPOSE) rm --force --stop -v
	IMAGE_TAG=${IMAGE_TAG} \
	$(DOCKER_COMPOSE) up -d --force-recreate --remove-orphans --build

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	$(DOCKER_COMPOSE) rm --force --stop -v
