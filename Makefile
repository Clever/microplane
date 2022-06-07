include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

.PHONY: test build release $(PKGS)
PKGS := $(shell go list ./... | grep -v vendor)
VERSION := $(shell head -n 1 VERSION)
EXECUTABLE := mp
$(eval $(call golang-version-check,1.17))

all: test build release

test: $(PKGS)
	@echo -e "\nAll done."

$(PKGS): golang-test-all-deps
	$(call golang-test-all,$@)

build:
	@go build -ldflags "-X main.version=$(VERSION)" -o ./bin/$(EXECUTABLE)

release:
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o="$@/$(EXECUTABLE)-$(VERSION)-linux-amd64"
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o="$@/$(EXECUTABLE)-$(VERSION)-darwin-amd64"

install_deps:
	go install github.com/GeertJohan/fgt@262f7b11eec07dc7b147c44641236f3212fee89d
	go install golang.org/x/lint/golint@738671d3881b9731cc63024d5d88cf28db875626
	go mod vendor
