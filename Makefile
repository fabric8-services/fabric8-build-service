PROJECT_NAME=fabric8-build
PACKAGE_NAME:=github.com/fabric8-services/$(PROJECT_NAME)
CUR_DIR=$(shell pwd)
TMP_PATH=$(CUR_DIR)/tmp
INSTALL_PREFIX=$(CUR_DIR)/bin
VENDOR_DIR=vendor
SOURCE_DIR ?= .
SOURCES := $(shell find $(SOURCE_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)
DESIGN_DIR=design
DESIGNS := $(shell find $(SOURCE_DIR)/$(DESIGN_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)
ALL_PKGS_EXCLUDE_PATTERN = 'vendor\|app\|tool\/cli\|design\|client\|test'
LDFLAGS=-ldflags "-X ${PACKAGE_NAME}/app.Commit=${COMMIT} -X ${PACKAGE_NAME}/app.BuildTime=${BUILD_TIME}"

# Paths common between OS
BINARY_SERVER_BIN=$(INSTALL_PREFIX)/fabric8-build
GOAGEN_BIN=$(VENDOR_DIR)/github.com/goadesign/goa/goagen/goagen
GO_BINDATA_DIR=$(VENDOR_DIR)/github.com/jteeuwen/go-bindata/go-bindata/
GO_BINDATA_BIN=$(GO_BINDATA_DIR)/go-bindata
FRESH_BIN=$(VENDOR_DIR)/github.com/chmouel/fresh/fresh
EXTRA_PATH=$(shell dirname $(GO_BINDATA_BIN))
GOCOV_BIN=$(VENDOR_DIR)/github.com/axw/gocov/gocov/gocov
GOCOVMERGE_BIN=$(VENDOR_DIR)/github.com/wadey/gocovmerge/gocovmerge
GOLINT_DIR=$(VENDOR_DIR)/github.com/golang/lint/golint
GOLINT_BIN=$(GOLINT_DIR)/golint
GOCYCLO_DIR=$(VENDOR_DIR)/github.com/fzipp/gocyclo
GOCYCLO_BIN=$(GOCYCLO_DIR)/gocyclo
GIT_BIN_NAME:=git
GO_BIN_NAME:=go
DEP_BIN_NAME:=dep

# by default use docker for compatibily and buildah/podman on Linux
CONTAINER_RUN := docker

# DB Container
DB_CONTAINER_NAME = db-build
DB_CONTAINER_PORT = 5433
DB_CONTAINER_IMAGE = registry.centos.org/postgresql/postgresql:9.6

# Auth
AUTH_CONTAINER_NAME = auth
AUTH_CONTAINER_PORT = 8089
AUTH_CONTAINER_IMAGE = quay.io/openshiftio/fabric8-services-fabric8-auth:latest

AUTH_DB_CONTAINER_NAME = db-auth
AUTH_DB_CONTAINER_IMAGE = $(DB_CONTAINER_IMAGE)

# Env
ENV_CONTAINER_NAME = f8env
ENV_CONTAINER_PORT = 8080
ENV_CONTAINER_IMAGE = quay.io/openshiftio/fabric8-services-fabric8-env:latest

ENV_DB_CONTAINER_NAME = db-env
ENV_DB_CONTAINER_IMAGE = $(DB_CONTAINER_IMAGE)

# By default reduce the amount of log output from tests
F8_LOG_LEVEL ?= error

# declares variable that are OS-sensitive
SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
include $(SELF_DIR)/.make/Makefile.lnx
endif

# This is a fix for a non-existing user in passwd file when running in a docker
# container and trying to clone repos of dependencies
GIT_COMMITTER_NAME ?= "Chmouel Boudjnah"
GIT_COMMITTER_EMAIL ?= "chmouel@chmouel.com"
export GIT_COMMITTER_NAME
export GIT_COMMITTER_EMAIL

COMMIT=$(shell git rev-parse HEAD 2>/dev/null)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
COMMIT := $(COMMIT)-dirty
endif
BUILD_TIME=`date -u '+%Y-%m-%dT%H:%M:%SZ'`

.DEFAULT_GOAL := help

# Call this function with $(call log-info,"Your message")
define log-info =
@echo "INFO: $(1)"
endef

# -------------------------------------------------------------------
# Container build
# -------------------------------------------------------------------
BUILD_DIR = bin
REGISTRY_URI = quay.io
REGISTRY_NS = fabric8-services
REGISTRY_IMAGE = ${PROJECT_NAME}

ifeq ($(TARGET),rhel)
	REGISTRY_URL_IMAGE := ${REGISTRY_URI}/openshiftio/rhel-${REGISTRY_NS}-${REGISTRY_IMAGE}
	CONTAINERFILE := ./.make/Dockerfile.rhel
else
	REGISTRY_URL_IMAGE := ${REGISTRY_URI}/openshiftio/${REGISTRY_NS}-${REGISTRY_IMAGE}
	CONTAINERFILE := ./.make/Dockerfile
endif

$(BUILD_DIR):
	mkdir $(BUILD_DIR)

.PHONY: build-linux $(BUILD_DIR)
build-linux: prebuild-check deps generate ## Builds the Linux binary for the container image into bin/ folder
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)

.PHONY: image
image: clean-artifacts build-linux ## Build the container image
ifeq ($(UNAME_S),Linux)
		buildah bud -t $(REGISTRY_URL_IMAGE) -f $(CONTAINERFILE) .
else
		docker build -t $(REGISTRY_URL_IMAGE) -f $(CONTAINERFILE) .
endif


# -------------------------------------------------------------------
# Unittest
# -------------------------------------------------------------------
.PHONY: test-unit
test-unit: prebuild-check $(SOURCES) generate ## Runs the unit tests and WITHOUT producing coverage files for each package.
	$(call log-info,"Running test: $@")
	$(eval TEST_PACKAGES:=$(shell go list ./... | grep -v $(ALL_PKGS_EXCLUDE_PATTERN)))
	F8_RESOURCE_UNIT_TEST=1 F8_RESOURCE_DATABASE=1 F8_DEVELOPER_MODE_ENABLED=1 \
	F8_LOG_LEVEL=$(F8_LOG_LEVEL) \
	go test -v $(GO_TEST_VERBOSITY_FLAG) $(TEST_PACKAGES)

.PHONY: coverage
coverage: prebuild-check deps $(SOURCES) ## Run coverage
	$(call log-info,"Running coverage: $@")
	$(eval TEST_PACKAGES:=$(shell go list ./... | grep -v $(ALL_PKGS_EXCLUDE_PATTERN)))
	@cd $(VENDOR_DIR)/github.com/haya14busa/goverage && go build
	F8_POSTGRES_PORT=$(DB_CONTAINER_PORT) \
	F8_DEVELOPER_MODE_ENABLED=1 F8_RESOURCE_UNIT_TEST=1 \
	F8_LOG_LEVEL=$(F8_LOG_LEVEL) F8_RESOURCE_DATABASE=1 \
	./vendor/github.com/haya14busa/goverage/goverage -v -coverprofile=tmp/coverage.out $(TEST_PACKAGES)
	sed -i~ '/\/sqlbindata.go:/d' tmp/coverage.out
	@go tool cover -func tmp/coverage.out

# -------------------------------------------------------------------
# help!
# -------------------------------------------------------------------
.PHONY: help
help: ## Prints this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

# -------------------------------------------------------------------
# required tools
# -------------------------------------------------------------------

# Find all required tools:
GIT_BIN := $(shell command -v $(GIT_BIN_NAME) 2> /dev/null)
DEP_BIN_DIR := $(TMP_PATH)/bin
DEP_BIN := $(DEP_BIN_DIR)/$(DEP_BIN_NAME)
DEP_VERSION=v0.4.1
GO_BIN := $(shell command -v $(GO_BIN_NAME) 2> /dev/null)

$(INSTALL_PREFIX):
	mkdir -p $(INSTALL_PREFIX)
$(TMP_PATH):
	mkdir -p $(TMP_PATH)

.PHONY: prebuild-check
prebuild-check: $(TMP_PATH) $(INSTALL_PREFIX)
# Check that all tools where found
ifndef GIT_BIN
	$(error The "$(GIT_BIN_NAME)" executable could not be found in your PATH)
endif
ifndef DEP_BIN
	$(error The "$(DEP_BIN_NAME)" executable could not be found in your PATH)
endif
ifndef GO_BIN
	$(error The "$(GO_BIN_NAME)" executable could not be found in your PATH)
endif

# -------------------------------------------------------------------
# deps
# -------------------------------------------------------------------
$(DEP_BIN_DIR):
	mkdir -p $(DEP_BIN_DIR)

.PHONY: deps
deps: $(DEP_BIN) $(VENDOR_DIR) ## Download build dependencies.

# install dep in a the tmp/bin dir of the repo
$(DEP_BIN): $(DEP_BIN_DIR)
	@echo "Installing 'dep' $(DEP_VERSION) at '$(DEP_BIN_DIR)'..."
	mkdir -p $(DEP_BIN_DIR)
ifeq ($(UNAME_S),Darwin)
	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-darwin-amd64 -o $(DEP_BIN)
	@cd $(DEP_BIN_DIR) && \
	curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-darwin-amd64.sha256 -o $(DEP_BIN_DIR)/dep-darwin-amd64.sha256 && \
	echo "1544afdd4d543574ef8eabed343d683f7211202a65380f8b32035d07ce0c45ef  dep" > dep-darwin-amd64.sha256 && \
	shasum -a 256 --check dep-darwin-amd64.sha256
else
	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-linux-amd64 -o $(DEP_BIN)
	@cd $(DEP_BIN_DIR) && \
	echo "31144e465e52ffbc0035248a10ddea61a09bf28b00784fd3fdd9882c8cbb2315  dep" > dep-linux-amd64.sha256 && \
	sha256sum -c dep-linux-amd64.sha256
endif
	@chmod +x $(DEP_BIN)

$(VENDOR_DIR): Gopkg.toml
	@echo "checking dependencies with $(DEP_BIN_NAME)"
	@$(DEP_BIN) ensure -v

# -------------------------------------------------------------------
# Code format/check
# -------------------------------------------------------------------
GOFORMAT_FILES := $(shell find  . -name '*.go' | grep -vEf .gofmt_exclude)

.PHONY: check-go-format
check-go-format: prebuild-check deps ## Exists with an error if there are files whose formatting differs from gofmt's
	@gofmt -s -l ${GOFORMAT_FILES} 2>&1 \
		| tee /tmp/gofmt-errors \
		| read \
	&& echo "ERROR: These files differ from gofmt's style (run 'make format-go-code' to fix this):" \
	&& cat /tmp/gofmt-errors \
	&& exit 1 \
	|| true

# TODO(chmou): https://git.io/fxzkM
.PHONY: analyze-go-code
analyze-go-code: deps generate ## Run golangci analysis over the code.
	$(info >>--- RESULTS: GOLANGCI CODE ANALYSIS ---<<)
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@golangci-lint run

.PHONY: format-go-code
format-go-code: prebuild-check ## Formats any go file that differs from gofmt's style
	@gofmt -s -l -w ${GOFORMAT_FILES}

# -------------------------------------------------------------------
# support for running in dev mode
# -------------------------------------------------------------------
$(FRESH_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/chmouel/fresh && go build -v

# -------------------------------------------------------------------
# support for generating goa code
# -------------------------------------------------------------------
$(GOAGEN_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v

# -------------------------------------------------------------------
# support for generating bindatas
# -------------------------------------------------------------------
$(GO_BINDATA_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/jteeuwen/go-bindata/go-bindata && go build -v

# -------------------------------------------------------------------
# clean
# -------------------------------------------------------------------

# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

CLEAN_TARGETS += clean-artifacts
.PHONY: clean-artifacts
## Removes the ./bin directory.
clean-artifacts:
	-rm -rf $(INSTALL_PREFIX)

CLEAN_TARGETS += clean-object-files
.PHONY: clean-object-files
## Runs go clean to remove any executables or other object files.
clean-object-files:
	go clean ./...

CLEAN_TARGETS += clean-generated
.PHONY: clean-generated
## Removes all generated code.
clean-generated:
	-rm -rf ./app
	-rm -rf ./swagger/
	-rm -f ./migration/sqlbindata.go
	-rm -rf ./auth/client

CLEAN_TARGETS += clean-vendor
.PHONY: clean-vendor
## Removes the ./vendor directory.
clean-vendor:
	-rm -rf $(VENDOR_DIR)

CLEAN_TARGETS += clean-tmp
.PHONY: clean-tmp
## Removes the ./vendor directory.
clean-tmp:
	-rm -rf $(TMP_DIR)

# Keep this "clean" target here after all `clean-*` sub tasks
.PHONY: clean
clean: $(CLEAN_TARGETS) ## Runs all clean-* targets.

# -------------------------------------------------------------------
# run in dev mode
# -------------------------------------------------------------------
.PHONY: dev
dev: prebuild-check deps generate $(FRESH_BIN)  ## run the server locally
	F8_POSTGRES_PORT=$(DB_CONTAINER_PORT) F8_DEVELOPER_MODE_ENABLED=true $(FRESH_BIN)

# -------------------------------------------------------------------
# build the binary executable (to ship in prod)
# -------------------------------------------------------------------
.PHONY: build
build: prebuild-check deps generate ## Build the server
	go build -v $(LDFLAGS) -o $(BINARY_SERVER_BIN)

# Pack all migration SQL files into a compilable Go file
migration/sqlbindata.go: $(GO_BINDATA_BIN) $(wildcard migration/sql-files/*.sql)
	$(GO_BINDATA_BIN) \
		-o migration/sqlbindata.go \
		-pkg migration \
		-prefix migration/sql-files \
		-nocompress \
		migration/sql-files

app/controllers.go: $(DESIGNS) $(GOAGEN_BIN) $(VENDOR_DIR)
	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg ${PACKAGE_NAME}/app
	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} --pkg-path=github.com/fabric8-services/fabric8-common/goasupport/status --out app
	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} \
		--pkg-path=github.com/fabric8-services/fabric8-common/goasupport/jsonapi_errors_helpers --out app
	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-auth/design --notool --out auth --pkg client

.PHONY: generate
## Generate GOA sources. Only necessary after clean of if changed `design` folder.
generate: app/controllers.go migration/sqlbindata.go

.PHONY: regenerate
regenerate: clean-generated generate ## Runs the "clean-generated" and the "generate" target

.PHONY: print-env
print-env:
	$(foreach var,$(.VARIABLES),$(info $(var)="$($(var))"))

# -----------------------------
# Run into a container for unitests
# -----------------------------
.PHONY: container-run
container-run: container-run-local-postgres ## Runs all the container images

.PHONY: container-run-local-postgres
container-run-local-postgres: container-clean-postgres ## Runs db in container
	$(info >>--- Starting container $(DB_CONTAINER_NAME) ---<<)
	 @[[ "`$(CONTAINER_RUN) ps -q --filter 'name=$(DB_CONTAINER_NAME)'`xxx" == xxx ]] && \
		$(CONTAINER_RUN) run --name $(DB_CONTAINER_NAME) -e POSTGRESQL_ADMIN_PASSWORD=`sed -n '/postgres.password/ { s/.*: //;p ;}' config.yaml` \
		 -d -p $(DB_CONTAINER_PORT):5432 $(DB_CONTAINER_IMAGE) >/dev/null
	sleep 2 # sleep for a bit that it started

.PHONY: container-clean-postgres
container-clean-postgres:
	$(info >>--- Stopping container $(DB_CONTAINER_NAME) ---<<)
	@$(CONTAINER_RUN) rm -f $(DB_CONTAINER_NAME) 2>/dev/null || true

.PHONY: deploy-openshift-dev
deploy-openshift-dev: ## Deploy to an (already running) openshift environement
	$(info >>-- Running the whole thing in openshift)
	@./openshift/deploy-openshift-dev.sh


.PHONY: deploy-minishift
deploy-minishift-dev: build ## Deploy to a minishift environement
	$(info >>-- Running in minishift)
	eval `minishift docker-env` && eval `minishift oc-env` && \
		make image && \
		./openshift/deploy-openshift-dev.sh
