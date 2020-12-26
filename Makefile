PROGRAM    := gitlabfs
TARGET_DIR := ./bin
VERSION    := $(shell git describe --tags --always)

SRC := $(shell find -name '*.go' -not -name '*_test.go' -type f)
BIN := $(TARGET_DIR)/$(PROGRAM)

# tests
ALL_SRC := $(shell find -name '*.go' -type f)
COVER_PROFILE := $(TARGET_DIR)/coverage.out
COVER_REPORT  := $(TARGET_DIR)/coverage.html

# go
GO       := go
GOFMT    := $(GO) fmt
GORUN    := $(GO) run -race
GOCLEAN  := $(GO) clean
GOTEST   := $(GO) test
GOCOVER  := $(GO) tool cover
GOBUILD  := $(GO) build

# docker
DOCKER      := docker
DOCKERBUILD := $(DOCKER) build
DOCKERTAG   := $(DOCKER) tag
DOCKERPUSH  := $(DOCKER) push
DOCKER_TAGS := $(PROGRAM):$(VERSION)

.PHONY: all
all: build

.PHONY: help
help:
	@echo "clean        revert back to clean state"
	@echo "lint         lint and format the code"
	@echo "run          run without building"
	@echo "prepare      prepare environment for build"
	@echo "test         build and run short test suite"
	@echo "coverage     build and run full test suite and generate coverage report"
	@echo "build        build the binary"
	@echo "build-docker build the docker image"

# clean: revert back to initial state

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -r $(TARGET_DIR)

# lint: lint and format the code

.PHONY: lint
lint:
	$(GOFMT) ./...

# run: run without building with debug flags

.PHONY: run
run: lint
	$(GORUN) main.go

# prepare: prepare environment for build

$(TARGET_DIR):
	mkdir $(TARGET_DIR) &>/dev/null || true

.PHONY: prepare
prepare: $(TARGET_DIR)

# test: run short test suite

.PHONY: test
test: lint
	$(GOTEST) -short ./...

# coverage: run full test suite with coverage report

$(COVER_PROFILE): $(TARGET_DIR) $(ALL_SRC)
	$(GOTEST) -coverprofile=$(COVER_PROFILE) ./...

$(COVER_REPORT): $(TARGET_DIR) $(COVER_PROFILE)
	$(GOCOVER) -html=$(COVER_PROFILE) -o=$(COVER_REPORT)

.PHONY: coverage
coverage: lint $(COVER_REPORT)
	$(GOCOVER) -func=$(COVER_PROFILE)

# build: build the program for release

$(BIN): $(TARGET_DIR) $(SRC)
	$(GOBUILD) -o=$(BIN)

.PHONY: build
build: lint $(BIN)

# build-docker: build the docker image

.PHONY: build-docker
build-docker: $(BIN)
	$(DOCKERBUILD) -t $(PROGRAM) .
	$(foreach tag,$(DOCKER_TAGS),$(DOCKERTAG) $(PROGRAM) $(tag);)

