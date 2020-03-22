PROJ=oidc-auth-service
ORG_PATH=github.com/fezho
REPO_PATH=$(ORG_PATH)/$(PROJ)
export PATH := $(PWD)/bin:$(PATH)

GitSHA=`git rev-parse HEAD`
Date=`date "+%Y-%m-%d %Z %H:%M:%S"`
VERSION ?= $(shell ./scripts/git-version)

DOCKER_REPO=docker.pkg.github.com/fezho/oidc-auth-service/auth-service
DOCKER_IMAGE=$(DOCKER_REPO):$(VERSION)

$( shell mkdir -p bin )

user=$(shell id -u -n)
group=$(shell id -g -n)

export GOBIN=$(PWD)/bin

LD_FLAGS=" -X $(REPO_PATH)/version.Version=$(VERSION)"
LD_FLAGS=" \
    -X '${REPO_PATH}/version.GitSHA=${GitSHA}' \
    -X '${REPO_PATH}/version.Built=${Date}'   \
    -X '${REPO_PATH}/version.Version=${VERSION}'"


# Dependency versions
GOLANGCI_VERSION = 1.21.0

build: bin/auth-service

bin/auth-service:
	@go install -v -ldflags $(LD_FLAGS) $(REPO_PATH)/cmd/auth-service

.PHONY: release-binary
release-binary:
	go build -o /go/bin/auth-service -v -ldflags $(LD_FLAGS) $(REPO_PATH)/cmd/auth-service

test:
	@go test -v ./...

testrace:
	@go test -v --race ./...

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	bin/golangci-lint run

.PHONY: fix
fix: bin/golangci-lint ## Fix lint violations
	bin/golangci-lint run --fix

.PHONY: image
image:
	@docker build -t $(DOCKER_IMAGE) .

.PHONY: get
get:
	@go get ./...
	@go mod verify
	@go mod tidy

.PHONY: update
update:
	@go get -u -v all
	@go mod verify
	@go mod tidy

clean:
	@rm -rf bin/

testall: testrace

FORCE:

.PHONY: test testrace testall
