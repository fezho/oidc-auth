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

.PHONY: revendor
revendor:
	@go mod tidy -v
	@go mod vendor -v
	@go mod verify

test: bin/test/kube-apiserver bin/test/etcd
	@go test -v ./...

testrace: bin/test/kube-apiserver bin/test/etcd
	@go test -v --race ./...

export TEST_ASSET_KUBE_APISERVER=$(abspath bin/test/kube-apiserver)
export TEST_ASSET_ETCD=$(abspath bin/test/etcd)

bin/test/kube-apiserver:
	@mkdir -p bin/test
	curl -L https://storage.googleapis.com/k8s-c10s-test-binaries/kube-apiserver-$(shell uname)-x86_64 > bin/test/kube-apiserver
	chmod +x bin/test/kube-apiserver

bin/test/etcd:
	@mkdir -p bin/test
	curl -L https://storage.googleapis.com/k8s-c10s-test-binaries/etcd-$(shell uname)-x86_64 > bin/test/etcd
	chmod +x bin/test/etcd

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

.PHONY: docker-image
docker-image:
	#@sudo docker build -t $(DOCKER_IMAGE) .
	@docker build -t $(DOCKER_IMAGE) .

clean:
	@rm -rf bin/

testall: testrace

FORCE:

.PHONY: test testrace testall
