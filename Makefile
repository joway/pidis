PROJECT_NAME = pikv
PACKAGES ?= $(shell go list ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GOFMT ?= gofmt "-s"
VERSION := v$(shell cat VERSION.go | grep -o -e '[0-9].[0-9].[0-9]')

.PHONY: all
all: install protoc build

.PHONY: install
install:
	@echo ">> install dependence"
	@exec ./bin/dep.sh

.PHONY: protoc
protoc:
	@echo ">> gen protobuf"
	@protoc --go_out=plugins=grpc:. ./rpc/**/*.proto

.PHONY: release
release:
	@echo ">> release ${VERSION}"
	@git tag "${VERSION}"
	@git push origin "${VERSION}"

.PHONY: test
test:
	@echo ">> run test"
	@go test -race -coverprofile=coverage.txt -covermode=atomic -v ./...

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
	@CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o pikv

.PHONY: docker-build
docker-build:
	@echo ">> building docker image"
	@docker build -t "joway/${PROJECT_NAME}:${VERSION}" .

.PHONY: docker-push
docker-push:
	@echo ">> push docker image"
	@docker push "joway/${PROJECT_NAME}:${VERSION}" .
