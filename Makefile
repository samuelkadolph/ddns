GO ?= go
PACKAGES := cmd/ddns cmd/ddns-client internal/config internal/route53 internal/server
SOURCES := $(filter-out %test.go, $(shell find internal -name '*.go'))

BUILD_DATE ?= `date -u +"%Y-%m-%dT%H:%M:%SZ"`
COMPOSE_DEV ?= docker-compose.dev.yml
IMAGE_NAME ?= samuelkadolph/ddns:latest
VCS_REF ?= `git rev-parse --short HEAD`

default: clean test build

build: build/ddns build/ddns-client
.PHONY: build

build/ddns: $(SOURCES) $(wildcard cmd/ddns/*.go)
	$(GO) build -o build/ddns ./cmd/ddns

build/ddns-client: $(SOURCES) $(wildcard cmd/ddns-client/*.go)
	$(GO) build -o build/ddns-client ./cmd/ddns-client

clean:
	rm -rf build/*
.PHONY: clean

docker-build:
	docker build --build-arg BUILD_DATE=$(BUILD_DATE) --build-arg VCS_REF=$(VCS_REF) --tag $(IMAGE_NAME) .
.PHONY: docker-build

docker-down:
	docker compose -f $(COMPOSE_DEV) down
.PHONY: docker-down

docker-publish: docker-build
	docker push $(IMAGE_NAME)
.PHONY: docker-publish

docker-run:
	docker compose -f $(COMPOSE_DEV) up --build
.PHONY: docker-run

docker-up:
	docker compose -f $(COMPOSE_DEV) up --build --detach
.PHONY: docker-up

fmt:
	$(GO) fmt $(addprefix ./,$(PACKAGES))
.PHONY: fmt

get:
	go get ./cmd/ddns ./cmd/ddns-client
.PHONY: get

test:
	$(GO) test $(addprefix ./,$(PACKAGES))
.PHONY: test
