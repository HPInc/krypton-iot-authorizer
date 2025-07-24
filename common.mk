TARGET=krypton
SRC=.

GOCMD=go
GOBUILD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOIMPORTS=goimports -w .

BINARY_DIR=bin
BINARY_NAME=aws-iot-authorizer

GIT_COMMIT := $(shell git rev-list -1 HEAD)
BUILT_ON := $(shell hostname)
BUILD_DATE := $(shell date +%FT%T%z)

LDFLAGS := "-X main.gitCommitHash=$(GIT_COMMIT) -X main.builtAt=$(BUILD_DATE) -X main.builtBy=$(USER) -X main.builtOn=$(BUILT_ON)"

DOCKER_IMAGE=krypton-aws-iot-authorizer

HAS_GO_LINT:=$(shell command -v golint 2> /dev/null)
HAS_GO_IMPORTS:=$(shell command -v goimports 2>/dev/null)

TESTCLEAN=(make -C test clean)

GHCR=ghcr.io/hpinc
REPO=krypton

clean:
	$(GOCLEAN)
	find . -name "*~" -exec rm -f {} \;
	docker rmi -f $(DOCKER_IMAGE) 2>/dev/null 1>&2
	docker image prune -f 2>/dev/null 1>&2
	rm -rf $(BINARY_DIR)

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

gosec:
	gosec ./...

run-test:
	(make -C test test)

unit-test:
	go test ./...
	
imports: check_goimports
	$(GOIMPORTS)

tag:
	docker tag $(DOCKER_IMAGE) $(GHCR)/$(REPO)/$(DOCKER_IMAGE)

push: tag
	docker push $(GHCR)/$(REPO)/$(DOCKER_IMAGE)

lint:
	./tools/run_linter.sh

check_golint:
		@which golint >/dev/null 2>&1 || (echo "golint not found";exit 1)

check_goimports:
		@which goimports >/dev/null 2>&1 || (echo "goimports not found";exit 1)
