REGISTRY?=steblynskyi-docker.jfrog.io
PKG_NAME?=go-serve
ARCH?=amd64
# Release variables
# ------------------
GIT_COMMIT?=$(shell git rev-parse "HEAD^{commit}" 2>/dev/null)
GIT_SHORT_COMMIT?=$(shell git rev-parse --short "HEAD^{commit}" 2>/dev/null)
GIT_TAG?=$(shell git describe --exact-match --tags "$(GIT_SHORT_COMMIT)" 2>/dev/null || echo "$(GIT_SHORT_COMMIT)")
BUILD_DATE:=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

VERSION := $(GIT_TAG)

LDFLAGS:=-X main.GitVersion=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)

.PHONY: go-serve
go-serve:
	@echo "Building go-serve"
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -mod=readonly -ldflags "$(LDFLAGS)" -o $(PKG_NAME)

.PHONY: container
container:
	@echo "Building container image"
	@echo $(PKG_NAME) $(VERSION) $(GIT_COMMIT) $(BUILD_DATE)
	@docker build -t $(REGISTRY)/$(PKG_NAME):$(VERSION) --build-arg ARCH=$(ARCH) --build-arg GIT_TAG=$(VERSION) --build-arg GIT_COMMIT=$(GIT_COMMIT) .
.PHONY: push
push:
	@echo "Pushing container image"
	@docker push $(REGISTRY)/$(PKG_NAME):$(VERSION)