# This is the default Clever Golang Makefile.
# It is stored in the dev-handbook repo, github.com/Clever/dev-handbook
# Please do not alter this file directly.
GOLANG_MK_VERSION := 1.3.0

SHELL := /bin/bash
SYSTEM := $(shell uname -a | cut -d" " -f1 | tr '[:upper:]' '[:lower:]')
.PHONY: golang-test-deps golang-ensure-curl-installed

# set timezone to UTC for golang to match circle and deploys
export TZ=UTC

# go build flags for use across all commands which accept them
export GOFLAGS := -mod=vendor $(GOFLAGS)

# if the gopath includes several directories, use only the first
GOPATH=$(shell echo $$GOPATH | cut -d: -f1)

# This block checks and confirms that the proper Go toolchain version is installed.
# It uses ^ matching in the semver sense -- you can be ahead by a minor
# version, but not a major version (patch is ignored).
# arg1: golang version
define golang-version-check
_ := $(if  \
		$(shell  \
			expr >/dev/null  \
				`go version | cut -d" " -f3 | cut -c3- | cut -d. -f2 | sed -E 's/beta[0-9]+//'`  \
				\>= `echo $(1) | cut -d. -f2`  \
				\&  \
				`go version | cut -d" " -f3 | cut -c3- | cut -d. -f1`  \
				= `echo $(1) | cut -d. -f1`  \
			&& echo 1),  \
		@echo "",  \
		$(error must be running Go version ^$(1) - you are running $(shell go version | cut -d" " -f3 | cut -c3-)))
endef

# FGT is a utility that exits with 1 whenever any stderr/stdout output is recieved.
# We pin its version since its a simple tool that does its job as-is;
# so we're defended against it breaking or changing in the future.
FGT := $(GOPATH)/bin/fgt
$(FGT):
	go install -mod=readonly github.com/GeertJohan/fgt@262f7b11eec07dc7b147c44641236f3212fee89d

golang-ensure-curl-installed:
	@command -v curl >/dev/null 2>&1 || { echo >&2 "curl not installed. Please install curl."; exit 1; }

# Golint is a tool for linting Golang code for common errors.
# We pin its version because an update could add a new lint check which would make
# previously passing tests start failing without changing our code.
# this package is deprecated and frozen
# Infra recomendation is to eventaully move to https://github.com/golangci/golangci-lint so don't fail on linting error for now
GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go install -mod=readonly golang.org/x/lint/golint@738671d3881b9731cc63024d5d88cf28db875626

# golang-fmt-deps requires the FGT tool for checking output
golang-fmt-deps: $(FGT)

# golang-fmt checks that all golang files in the pkg are formatted correctly.
# arg1: pkg path
define golang-fmt
@echo "FORMATTING $(1)..."
@PKG_PATH=$$(go list -f '{{.Dir}}' $(1)); $(FGT) gofmt -l=true $${PKG_PATH}/*.go
endef

# golang-lint-deps requires the golint tool for golang linting.
golang-lint-deps: $(GOLINT)

# golang-lint calls golint on all golang files in the pkg.
# arg1: pkg path
define golang-lint
@echo "LINTING $(1)..."
@PKG_PATH=$$(go list -f '{{.Dir}}' $(1)); find $${PKG_PATH}/*.go -type f | grep -v gen_ | xargs $(GOLINT)
endef

# golang-lint-deps-strict requires the golint tool for golang linting.
golang-lint-deps-strict: $(GOLINT) $(FGT)

# golang-test-deps is here for consistency
golang-test-deps:

# golang-test uses the Go toolchain to run all tests in the pkg.
# arg1: pkg path
define golang-test
@echo "TESTING $(1)..."
@go test -v $(1)
endef

# golang-test-strict-deps is here for consistency
golang-test-strict-deps:

# golang-test-strict uses the Go toolchain to run all tests in the pkg with the race flag
# arg1: pkg path
define golang-test-strict
@echo "TESTING $(1)..."
@go test -v -race $(1)
endef

# golang-test-strict-cover-deps is here for consistency
golang-test-strict-cover-deps:

# golang-test-strict-cover uses the Go toolchain to run all tests in the pkg with the race and cover flag.
# appends coverage results to coverage.txt
# arg1: pkg path
define golang-test-strict-cover
@echo "TESTING $(1)..."
@go test -v -race -cover -coverprofile=profile.tmp -covermode=atomic $(1)
@if [ -f profile.tmp ]; then \
  cat profile.tmp | tail -n +2 >> coverage.txt; \
  rm profile.tmp; \
fi;
endef

# golang-vet-deps is here for consistency
golang-vet-deps:

# golang-vet uses the Go toolchain to vet all the pkg for common mistakes.
# arg1: pkg path
define golang-vet
@echo "VETTING $(1)..."
@go vet $(1)
endef

# golang-test-all-deps installs all dependencies needed for different test cases.
golang-test-all-deps: golang-fmt-deps golang-lint-deps golang-test-deps golang-vet-deps

# golang-test-all calls fmt, lint, vet and test on the specified pkg.
# arg1: pkg path
define golang-test-all
$(call golang-fmt,$(1))
$(call golang-lint,$(1))
$(call golang-vet,$(1))
$(call golang-test,$(1))
endef

# golang-test-all-strict-deps: installs all dependencies needed for different test cases.
golang-test-all-strict-deps: golang-fmt-deps golang-lint-deps-strict golang-test-strict-deps golang-vet-deps

# golang-test-all-strict calls fmt, lint, vet and test on the specified pkg with strict
# requirements that no errors are thrown while linting.
# arg1: pkg path
define golang-test-all-strict
$(call golang-fmt,$(1))
$(call golang-lint,$(1))
$(call golang-vet,$(1))
$(call golang-test-strict,$(1))
endef

# golang-test-all-strict-cover-deps: installs all dependencies needed for different test cases.
golang-test-all-strict-cover-deps: golang-fmt-deps golang-lint-deps-strict golang-test-strict-cover-deps golang-vet-deps

# golang-test-all-strict-cover calls fmt, lint, vet and test on the specified pkg with strict and cover
# requirements that no errors are thrown while linting.
# arg1: pkg path
define golang-test-all-strict-cover
$(call golang-fmt,$(1))
$(call golang-lint,$(1))
$(call golang-vet,$(1))
$(call golang-test-strict-cover,$(1))
endef

# golang-build: builds a golang binary
# arg1: pkg path
# arg2: executable name
define golang-build
@echo "BUILDING $(2)..."
@CGO_ENABLED=0 go build -o bin/$(2) $(1);
endef

# golang-debug-build: builds a golang binary with debugging capabilities
# arg1: pkg path
# arg2: executable name
define golang-debug-build
@echo "BUILDING $(2) FOR DEBUG..."
@CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/$(2) $(1);
endef

# golang-cgo-build: builds a golang binary with CGO
# arg1: pkg path
# arg2: executable name
define golang-cgo-build
@echo "BUILDING $(2) WITH CGO ..."
@CGO_ENABLED=1 go build -installsuffix cgo -o bin/$(2) $(1);
endef

# golang-setup-coverage: set up the coverage file
golang-setup-coverage:
	@echo "mode: atomic" > coverage.txt

# golang-update-makefile downloads latest version of golang.mk
golang-update-makefile:
	@wget https://raw.githubusercontent.com/Clever/dev-handbook/master/make/golang-v1.mk -O /tmp/golang.mk 2>/dev/null
	@if ! grep -q $(GOLANG_MK_VERSION) /tmp/golang.mk; then cp /tmp/golang.mk golang.mk && echo "golang.mk updated"; else echo "golang.mk is up-to-date"; fi
