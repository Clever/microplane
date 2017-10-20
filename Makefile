include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: test build release $(PKGS)
PKGS := $(shell go list ./... | grep -v vendor)
VERSION := $(shell head -n 1 VERSION)
EXECUTABLE := mp
$(eval $(call golang-version-check,1.9))

$(GOPATH)/bin/dep:
	@go get github.com/golang/dep
	@go install github.com/golang/dep/...

all: test build release

test: $(PKGS)
	@echo -e "\nAll done."

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

build:
	@go build -ldflags "-X main.version=$(VERSION)" -o ./bin/$(EXECUTABLE)

install_deps: $(GOPATH)/bin/dep
	@$(GOPATH)/bin/dep ensure

release:
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o="$@/$(EXECUTABLE)-$(VERSION)-linux-amd64"
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o="$@/$(EXECUTABLE)-$(VERSION)-darwin-amd64"
