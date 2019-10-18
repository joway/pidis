PROJECT_NAME = pikv
PACKAGES ?= $(shell go list ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GOFMT ?= gofmt "-s"
GO_BUILD_TAGS = -a -ldflags '-extldflags "-static"'
VERSION := v$(shell cat VERSION.go | grep -o -e '[0-9].[0-9].[0-9]')

.PHONY: all
all: install build

.PHONY: pre-commit
pre-commit: protoc fmt

.PHONY: install
install:
	@echo ">> install dependence"
	@exec ./tools/dep.sh

.PHONY: protoc
protoc:
	@echo ">> gen protobuf"
	@protoc --go_out=plugins=grpc:. ./proto/*.proto

.PHONY: release
release:
	@echo ">> release ${VERSION}"
	@git tag "${VERSION}"
	@git push origin "${VERSION}"

.PHONY: test
test:
	@echo ">> run test"
	@go test -p 1 -race -coverprofile=coverage.txt -covermode=atomic -v ./...

.PHONY: e2e
e2e:
	@echo ">> run e2e test"
	@go test -p 1 -race -coverprofile=coverage.txt -covermode=atomic -v ./e2e/...

.PHONY: fmt
fmt:
	@echo ">> formatting code"
	@$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	# get all go files and run go fmt on them
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: build
build: $(GOFILES)
	@echo ">> building binaries"
	@CGO_ENABLED=0 go build $(GO_BUILD_TAGS) -o bin/pikv cmd/pikv/*

.PHONY: docker-build
docker-build:
	@echo ">> building docker image"
	@docker build -t "joway/${PROJECT_NAME}:${VERSION}" .

.PHONY: docker-push
docker-push:
	@echo ">> push docker image"
	@docker push "joway/${PROJECT_NAME}:${VERSION}"
	@docker tag "joway/${PROJECT_NAME}:${VERSION}" "joway/${PROJECT_NAME}:latest"
	@docker push "joway/${PROJECT_NAME}:latest"
