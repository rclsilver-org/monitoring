BINARY = daemon
SOURCE_FILES = ./main.go

MAIN_PKG    = $(shell go list)
VERSION_PKG = ${MAIN_PKG}/version

VERSION    ?= $(shell generate-version.sh)
LAST_COMMIT = $(shell git rev-parse HEAD)

TEST_LOCATION ?= ./...
TEST_CMD       = go test -v -race -cover

LD_FLAGS = -ldflags "-X ${VERSION_PKG}.commit=${LAST_COMMIT} -X ${VERSION_PKG}.version=${VERSION}"

DOCKER_REPOSITORY ?= rclsilver
DOCKER_IMAGE = monitoring-daemon
DOCKER_TAG ?= latest

all: $(BINARY)

$(BINARY): $(SOURCE_FILES) go.mod
	CGO_ENABLED=0 go build $(LD_FLAGS) -o $@ $(SOURCE_FILES)

.PHONY: test
test:
	docker-compose up -d postgresql-test
	$(TEST_CMD) $(COVER_OPTS) $(TEST_LOCATION)

docker:
	docker build -t $(DOCKER_REPOSITORY)/$(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: clean
clean:
	rm -f $(BINARY)