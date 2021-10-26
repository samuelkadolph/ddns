SOURCES := $(filter-out %test.go, $(shell find internal -name '*.go'))
PACKAGES := cmd/ddns cmd/ddns-client internal/config internal/route53 internal/server

BUILD_DATE ?= `date -u +"%Y-%m-%dT%H:%M:%SZ"`
GO ?= go
IMAGE_NAME ?= samuelkadolph/ddns:latest
VCS_REF ?= `git rev-parse --short HEAD`

default: build

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

docker-publish: docker-build
	docker push $(IMAGE_NAME)

fmt:
	$(GO) fmt $(addprefix ./,$(PACKAGES))
.PHONY: fmt

test:
	$(GO) test $(addprefix ./,$(PACKAGES))
.PHONY: test
