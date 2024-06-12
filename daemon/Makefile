BINARY = monitoring-daemon
SOURCE_FILES = $(shell find . -type f -name '*.go' -not -name '*_test.go')

MAIN_PKG    = $(shell go list)
VERSION_PKG = ${MAIN_PKG}/version

VERSION    ?= $(shell ./generate-version.sh)
LAST_COMMIT = $(shell git rev-parse HEAD)

DIST_DIR = ./dist

TEST_LOCATION ?= ./...
TEST_CMD       = go test -v -race -cover

LD_FLAGS = -ldflags "-w -s -X ${VERSION_PKG}.commit=${LAST_COMMIT} -X ${VERSION_PKG}.version=${VERSION}"

all: monitoring-daemon-$(shell go env GOOS)-$(shell go env GOARCH)

compile: $(BINARY)

$(BINARY): $(BINARY)-linux-amd64 $(BINARY)-linux-mips64

$(BINARY)-linux-amd64: $(DIST_DIR)/$(BINARY)-linux-amd64

$(BINARY)-linux-mips64: $(DIST_DIR)/$(BINARY)-linux-mips64

$(DIST_DIR)/$(BINARY)-linux-amd64: $(SOURCE_FILES) go.mod
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LD_FLAGS) -o $@ ./main.go

$(DIST_DIR)/$(BINARY)-linux-mips64: $(SOURCE_FILES) go.mod
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build $(LD_FLAGS) -o $@ ./main.go

.PHONY: test
test:
	$(TEST_CMD) $(COVER_OPTS) $(TEST_LOCATION)

.PHONY: clean
clean:
	rm -f $(DIST_DIR)/$(BINARY)-amd64 $(DIST_DIR)/$(BINARY)-mips64
