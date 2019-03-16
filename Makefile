export GO111MODULE=on
BUILD_VERBOSE=1
VERSION = $(shell git describe --dirty --tags --always)
REPO = github.com/jwmatthews/overlook
BUILD_PATH = $(REPO)
PKGS = $(shell go list ./... | grep -v /vendor/)
BINARY_NAME=overlook

export CGO_ENABLED:=1
ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

BIN_DIR := $(GOPATH)/bin

.PHONY: all test format dep clean lint build install release_x86_64 release
all: format build

run_report: build/${BINARY_NAME}
	pushd . && cd build && ./${BINARY_NAME} report && popd

run_watch: build/${BINARY_NAME}
	pushd . && cd build && ./${BINARY_NAME} watch && popd

${GOPATH}/bin/golangci-lint:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

format:
	$(Q)go fmt $(PKGS)

dep:
	$(Q)dep ensure -v

dep-update:
	$(Q)dep ensure -update -v

test:
	go test -timeout 30s $(REPO)/pkg/...

lint: ${GOPATH}/bin/golangci-lint
	${GOPATH}/bin/golangci-lint run

clean:
	$(Q)rm build/${BINARY_NAME}*

install:
	$(Q)go install $(BUILD_PATH)

release_x86_64 := \
	build/${BINARY_NAME}-$(VERSION)-x86_64-linux-gnu \
	build/${BINARY_NAME}-$(VERSION)-x86_64-apple-darwin

release: clean $(release_x86_64) $(release_x86_64:=.asc)

build/${BINARY_NAME}-%-x86_64-linux-gnu: GOARGS = GOOS=linux GOARCH=amd64
build/${BINARY_NAME}-%-x86_64-apple-darwin: GOARGS = GOOS=darwin GOARCH=amd64

build/${BINARY_NAME}: build

build: lint
	$(Q)$(GOARGS) go build -o build/${BINARY_NAME} $(BUILD_PATH)

build/%.asc:
	$(Q){ \
	default_key=$$(gpgconf --list-options gpg | awk -F: '$$1 == "default-key" { gsub(/"/,""); print toupper($$10)}'); \
	git_key=$$(git config --get user.signingkey | awk '{ print toupper($$0) }'); \
	if [ "$${default_key}" = "$${git_key}" ]; then \
		gpg --output $@ --detach-sig build/$*; \
		gpg --verify $@ build/$*; \
	else \
		echo "git and/or gpg are not configured to have default signing key $${default_key}"; \
		exit 1; \
	fi; \
	}







