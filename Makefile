-include dev.env

## Set all the environment variables here
# Docker Registry
DOCKER_REGISTRY ?= docker.io/rocm

# Build Container environment
DOCKER_BUILDER_TAG ?= v1.1
BUILD_BASE_IMAGE ?= ubuntu:22.04
BUILD_CONTAINER ?= $(DOCKER_REGISTRY)/device-metrics-exporter-build:$(DOCKER_BUILDER_TAG)

# Exporter container environment
EXPORTER_IMAGE_TAG ?= latest
EXPORTER_IMAGE_NAME ?= device-metrics-exporter
EXPORTER_SRIOV_BASE_IMAGE ?= registry.access.redhat.com/ubi9/ubi-minimal:9.6
EXPORTER_SRIOV_IMAGE_NAME ?= device-metrics-exporter-sriov
RHEL_BASE_MIN_IMAGE ?= registry.access.redhat.com/ubi9/ubi-minimal:9.6
AZURE_BASE_IMAGE ?= mcr.microsoft.com/azurelinux/base/core:3.0

# Test runner container environment
TESTRUNNER_IMAGE_TAG ?= latest
TESTRUNNER_IMAGE_NAME ?= test-runner
TEST_RUNNER_RHEL_BASE_IMAGE ?= registry.access.redhat.com/ubi9/ubi-minimal:9.5

# External repo builders
GPUAGENT_BASE_IMAGE ?= ubuntu:22.04
GPUAGENT_BUILDER_IMAGE ?= gpuagent-builder:v1
AMDSMI_BASE_IMAGE ?= registry.access.redhat.com/ubi9/ubi:9.6
AMDSMI_BASE_UBUNTU22 ?= ubuntu:22.04
AMDSMI_BASE_UBUNTU24 ?= ubuntu:24.04
AMDSMI_BASE_AZURE ?= mcr.microsoft.com/azurelinux/base/core:3.0
ROCPROFILER_BASE_UBUNTU22 ?= ubuntu:22.04
AMDSMI_BUILDER_IMAGE ?= amdsmi-builder:rhel9
AMDSMI_BUILDER_UB22_IMAGE ?= amdsmi-builder:ub22
AMDSMI_BUILDER_UB24_IMAGE ?= amdsmi-builder:ub24
AMDSMI_BUILDER_AZURE_IMAGE ?= amdsmi-builder:azure
GIMSMI_BUILDER_IMAGE ?= gimsmi-builder:rhel9
GIMSMI_BUILDER_UB22_IMAGE ?= gimsmi-builder:ub22
GIMSMI_BUILDER_UB24_IMAGE ?= gimsmi-builder:ub24
ROCPROFILER_BUILDER_IMAGE ?= rocprofiler-builder:ub22

# export environment variables used across project
export DOCKER_REGISTRY
export BUILD_CONTAINER
export BUILD_BASE_IMAGE
export EXPORTER_IMAGE_NAME
export EXPORTER_SRIOV_BASE_IMAGE
export EXPORTER_SRIOV_IMAGE_NAME
export EXPORTER_IMAGE_TAG

# testrunner base images
export TESTRUNNER_IMAGE_NAME
export TESTRUNNER_IMAGE_TAG

# exporter base container images
export TEST_RUNNER_RHEL_BASE_IMAGE
export RHEL_BASE_MIN_IMAGE
export AZURE_BASE_IMAGE

# asset builder base images and tags
export AMDSMI_BASE_IMAGE
export AMDSMI_BASE_UBUNTU22
export AMDSMI_BASE_UBUNTU24
export AMDSMI_BASE_AZURE
export GPUAGENT_BUILDER_IMAGE
export ROCPROFILER_BASE_UBUNTU22

export AMDSMI_BUILDER_IMAGE
export AMDSMI_BUILDER_UB22_IMAGE
export AMDSMI_BUILDER_UB24_IMAGE
export AMDSMI_BUILDER_AZURE_IMAGE
export GPUAGENT_BASE_IMAGE
export ROCPROFILER_BUILDER_IMAGE

TO_GEN := pkg/amdgpu/proto pkg/exporter/proto pkg/amdnic/proto
TO_MOCK := pkg/amdgpu/mock pkg/exporter/scheduler
OUT_DIR := bin
CUR_USER:=$(shell whoami)
CUR_TIME:=$(shell date +%Y-%m-%d_%H.%M.%S)
CONTAINER_NAME:=${CUR_USER}_exporter-bld
CONTAINER_WORKDIR := /usr/src/github.com/ROCm/device-metrics-exporter

TOP_DIR := $(PWD)
GEN_DIR := $(TOP_DIR)/pkg/amdgpu/
MOCK_DIR := ${TOP_DIR}/pkg/amdgpu/mock_gen
HELM_CHARTS_DIR := $(TOP_DIR)/helm-charts
CONFIG_DIR := $(TOP_DIR)/example/
GOINSECURE='github.com, google.golang.org, golang.org'
GOFLAGS ='-buildvcs=false'
BUILD_DATE ?= $(shell date   +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT ?= $(shell git rev-list -1 HEAD --abbrev-commit)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
VERSION ?= $(if $(RELEASE),$(RELEASE),$(GIT_BRANCH))


KUBECONFIG ?= ~/.kube/config

# docs build settings
DOCS_DIR := ${TOP_DIR}/docs
BUILD_DIR := $(DOCS_DIR)/_build
HTML_DIR := $(BUILD_DIR)/html

# library branch to build amdsmi libraries for gpuagent
AMDSMI_BRANCH ?= rocm-7.0.0
AMDSMI_COMMIT ?= 37d158a
GIMSMI_BRANCH ?= mainline
GIMSMI_COMMIT ?= mainline/8.3.0.K

ROCM_VERSION ?= 7.0

export ${GOROOT}
export ${GOPATH}
export ${OUT_DIR}
export ${TOP_DIR}
export ${GOFLAGS}
export ${GOINSECURE}
export ${KUBECONFIG}
export ${AZURE_DOCKER_CONTAINER_IMG}
export ${BUILD_VER_ENV}
export ${AMDSMI_BRANCH}
export ${AMDSMI_COMMIT}
export ${GIMSMI_BRANCH}
export ${GIMSMI_COMMIT}

ASSETS_PATH :=${TOP_DIR}/assets
# 22.04 - jammy
# 24.04 - noble
UBUNTU_VERSION ?= jammy
UBUNTU_VERSION_NUMBER = 22.04
UBUNTU_LIBDIR = UBUNTU22
ifeq (${UBUNTU_VERSION}, noble)
UBUNTU_VERSION_NUMBER = 24.04
UBUNTU_LIBDIR = UBUNTU24
endif

# set version and run `make update-version` to all docs
PACKAGE_VERSION ?= "1.5.0"
ifneq (,$(findstring exporter,$(RELEASE)))
#remove prefix from main tag
DEBIAN_VERSION := $(shell echo "$(RELEASE)" | cut -c 10-)
else ifneq (,$(findstring v,$(RELEASE)))
#remove prefix for release tag
DEBIAN_VERSION := $(shell echo "$(RELEASE)" | sed 's/^.//')
else
#apt is only released until this version
DEBIAN_VERSION := "1.5.0"
endif
# Split DEBIAN_VERSION on '-' to get RPM version and release label
RPM_BUILD_VERSION := $(word 1,$(subst -, ,$(DEBIAN_VERSION)))
RPM_RELEASE_LABEL_TMP := $(word 2,$(subst -, ,$(DEBIAN_VERSION)))
RPM_RELEASE_LABEL := $(if $(RPM_RELEASE_LABEL_TMP),$(RPM_RELEASE_LABEL_TMP),0)
REL_IMAGE_TAG := $(subst $\",,v$(PACKAGE_VERSION))
HELM_VERSION := $(REL_IMAGE_TAG)

DOCS_DIR := $(TOP_DIR)/docs
DOCS_CONFIG_DIR := $(DOCS_DIR)/configuration/
DOCS_INSTALLATION_DIR := $(DOCS_DIR)/installation/
DOCS_INTEGRATION_DIR := $(DOCS_DIR)/integrations/

UPDATE_VERSION_TARGET_DIRS := $(DOCS_DIR)/configuration/ $(DOCS_DIR)/installation/ $(DOCS_DIR)/integrations/

export ${DEBIAN_VERSION}
export ${RPM_BUILD_VERSION}
export ${RPM_RELEASE_LABEL}

.PHONY: update-version
update-version:
	@echo "Replacing versions with $(PACKAGE_VERSION)..."
	@sed -i -e 's|version = .*|version = ${PACKAGE_VERSION}|' docs/conf.py
	@sed -i -e 's|tag:.*|tag: ${REL_IMAGE_TAG}|' helm-charts/values.yaml
	@sed -i -e 's|version:.*|version: ${HELM_VERSION}|' helm-charts/Chart.yaml
	@sed -i -e 's|appVersion:.*|appVersion: ${HELM_VERSION}|' helm-charts/Chart.yaml
	@sed -i -e 's|debian_version = .*|debian_version = ${DEBIAN_VERSION}|' docs/conf.py
	@for dir in $(UPDATE_VERSION_TARGET_DIRS); do \
		if [ -d $$dir ]; then \
			find $$dir -type f -exec sed -i -E "/Kubernetes/!s/v1+\.[0-9]+\.[0-9]+/v$(PACKAGE_VERSION)/g" {} +; \
			find $$dir -type f -exec sed -i -E "/Kubernetes/!s/1+\.[0-9]+\.[0-9]+/$(PACKAGE_VERSION)/g" {} +; \
		fi \
	done




TO_GEN_TESTRUNNER := pkg/testrunner/proto
GEN_DIR_TESTRUNNER := $(TOP_DIR)/pkg/testrunner/

include Makefile.build
include Makefile.compile
include Makefile.package

##################
# Makefile targets
#
##@ QuickStart
.PHONY: default
default: build-dev-container ## Quick start to build everything from docker shell container
	${MAKE} docker-compile

.PHONY: docker-shell
docker-shell:
	docker run --rm -it --privileged \
		--name ${CONTAINER_NAME} \
		-e "USER_NAME=$(shell whoami)" \
		-e "USER_UID=$(shell id -u)" \
		-e "USER_GID=$(shell id -g)" \
		-e "GIT_COMMIT=${GIT_COMMIT}" \
		-e "GIT_VERSION=${GIT_VERSION}" \
		-e "BUILD_DATE=${BUILD_DATE}" \
		-v $(CURDIR):$(CONTAINER_WORKDIR) \
		-w $(CONTAINER_WORKDIR) \
		$(BUILD_CONTAINER) \
		bash -c "cd $(CONTAINER_WORKDIR) && git config --global --add safe.directory $(CONTAINER_WORKDIR) && bash"

.PHONY: docker-compile
docker-compile:
	docker run --rm -it --privileged \
		--name ${CONTAINER_NAME} \
		-e "USER_NAME=$(shell whoami)" \
		-e "USER_UID=$(shell id -u)" \
		-e "USER_GID=$(shell id -g)" \
		-e "GIT_COMMIT=${GIT_COMMIT}" \
		-e "GIT_VERSION=${GIT_VERSION}" \
		-e "BUILD_DATE=${BUILD_DATE}" \
		-v $(CURDIR):$(CONTAINER_WORKDIR) \
		-w $(CONTAINER_WORKDIR) \
		$(BUILD_CONTAINER) \
		bash -c "cd $(CONTAINER_WORKDIR) && source ~/.bashrc && git config --global --add safe.directory $(CONTAINER_WORKDIR) && make all"

.PHONY: all
all:
	${MAKE} gen amdexporter metricutil amdtestrunner

.PHONY: gen
gen: gopkglist gen-test-runner
	@for c in ${TO_GEN}; do printf "\n+++++++++++++++++ Generating $${c} +++++++++++++++++\n"; PATH=$$PATH make -C $${c} GEN_DIR=$(GEN_DIR) || exit 1; done
	@for c in ${TO_MOCK}; do printf "\n+++++++++++++++++ Generating mock $${c} +++++++++++++++++\n"; PATH=$$PATH make -C $${c} MOCK_DIR=$(MOCK_DIR) GEN_DIR=$(GEN_DIR) || exit 1; done

.PHONY: gen-test-runner
gen-test-runner: gopkglist
	@for c in ${TO_GEN_TESTRUNNER}; do printf "\n+++++++++++++++++ Generating $${c} +++++++++++++++++\n"; PATH=$$PATH make -C $${c} GEN_DIR=$(GEN_DIR_TESTRUNNER) || exit 1; done

.PHONY:clean
clean: pkg-clean
	rm -rf bin
	rm -rf docker/obj
	rm -rf docker/*.tgz
	rm -rf docker/*.tar
	rm -rf docker/*.tar.gz
	rm -rf ${PKG_PATH}
	rm -rf build
	rm -rf helm-charts/*.tgz

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8)

HELMDOCS = $(shell pwd)/bin/helm-docs
.PHONY: helm-docs
helm-docs: ## Download helm-docs locally if necessary
	$(call go-get-tool,$(HELMDOCS),github.com/norwoodj/helm-docs/cmd/helm-docs@v1.12.0)
	$(HELMDOCS) -c $(shell pwd)/helm-charts/ -g $(shell pwd)/helm-charts -u --ignore-non-descriptions

# go-get-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
}
endef

EXCLUDE_PATTERN := "libamdsmi|gpuagent.sw|gpuagent.sw.nic|gpuagent.sw.nic.gpuagent"
GO_PKG := $(shell go list ./...  2>/dev/null | grep github.com/ROCm/device-metrics-exporter | egrep -v ${EXCLUDE_PATTERN})

GOFILES_NO_VENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
.PHONY: lint
lint: golangci-lint ## Run golangci-lint against code.
	@if [ `gofmt -l $(GOFILES_NO_VENDOR) | wc -l` -ne 0 ]; then \
		echo There are some malformed files, please make sure to run \'make fmt\'; \
		gofmt -l $(GOFILES_NO_VENDOR); \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run -v --timeout 5m0s

.PHONY: fmt
fmt:## Run go fmt against code.
	go fmt $(GO_PKG)

.PHONY: vet
vet: ## Run go vet against code.
	$(info +++ govet sources)
	go vet -source $(GO_PKG)

.PHONY: gopkglist
gopkglist:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install go.uber.org/mock/mockgen@v0.5.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/alta/protopatch/cmd/protoc-gen-go-patch@latest

amdexporter: metricsclient amdgpuhealth
	@echo "building amd metrics exporter"
	CGO_ENABLED=0 go build  -C cmd/exporter -ldflags "-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE} -X main.Publish=${DISABLE_DEBUG}" -o $(CURDIR)/bin/amd-metrics-exporter

amdtestrunner:
	@echo "building amd test runner"
	CGO_ENABLED=0 go build  -C cmd/testrunner -ldflags "-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}" -o $(CURDIR)/bin/amd-test-runner

metricutil:
	@echo "building metrics util"
	CGO_ENABLED=0 go build -C tools/metricutil -o $(CURDIR)/bin/metricutil

metricsclient:
	@echo "building metrics client"
	CGO_ENABLED=0 go build -C tools/metricsclient -o $(CURDIR)/bin/metricsclient

amdgpuhealth:
	@echo "building amd gpu health util"
	CGO_ENABLED=0 go build -C tools/amd-gpu-health -o $(CURDIR)/bin/amdgpuhealth

.PHONY: docker-cicd
docker-cicd: gen amdexporter
	echo "Building cicd docker for publish"
	${MAKE} -C docker docker-cicd TOP_DIR=$(CURDIR)
	${MAKE} -C docker docker-save TOP_DIR=$(CURDIR)

.PHONY: docker
docker: gen amdexporter
	${MAKE} -C docker TOP_DIR=$(CURDIR)
	${MAKE} -C docker docker-save TOP_DIR=$(CURDIR)

.PHONY: docker-sriov
docker-sriov: gen amdexporter
	echo "Building docker for sriov driver rhel9"
	${MAKE} -C docker docker-sriov TOP_DIR=$(CURDIR)
	${MAKE} -C docker docker-sriov-save TOP_DIR=$(CURDIR)

# for development we use ubuntu based
.PHONY: docker-sriov-ub22
docker-sriov-ub22: gen amdexporter
	echo "Building docker for sriov driver ub22"
	${MAKE} -C docker docker-sriov-ub22 TOP_DIR=$(CURDIR)
	${MAKE} -C docker docker-sriov-save TOP_DIR=$(CURDIR)

.PHONY: docker-mock
docker-mock: gen amdexporter
	${MAKE} -C docker TOP_DIR=$(CURDIR) MOCK=1 EXPORTER_IMAGE_NAME=$(EXPORTER_IMAGE_NAME)-mock
	${MAKE} -C docker docker-save TOP_DIR=$(CURDIR) EXPORTER_IMAGE_NAME=$(EXPORTER_IMAGE_NAME)-mock

.PHONY: docker-test-runner
docker-test-runner: gen-test-runner amdtestrunner
	${MAKE} -C docker/testrunner TOP_DIR=$(CURDIR) docker

.PHOHY: docker-test-runner-cicd
docker-test-runner-cicd: gen-test-runner amdtestrunner
	echo "Building test runner cicd docker for publish"
	${MAKE} -C docker/testrunner TOP_DIR=$(CURDIR) docker-cicd
	${MAKE} -C docker/testrunner TOP_DIR=$(CURDIR) docker-save

.PHONY: docker-azure
docker-azure: gen amdexporter
	${MAKE} -C docker azure TOP_DIR=$(CURDIR)
	${MAKE} -C docker docker-save TOP_DIR=$(CURDIR) DOCKER_CONTAINER_IMAGE=${EXPORTER_IMAGE_NAME}-${EXPORTER_IMAGE_TAG}-azure

.PHONY:checks
checks: gen vet lint

.PHONY: docker-publish
docker-publish:
	${MAKE} -C docker docker-publish TOP_DIR=$(CURDIR)

.PHONY: unit-test
unit-test:
	PATH=$$PATH LOGDIR=$(TOP_DIR)/ go test -v -cover -mod=vendor ./pkg/...

loadgpu:
	sudo modprobe amdgpu

mod:
	@echo "ignoring submodules gpuagent and libamdsmi"
	@touch ${TOP_DIR}/gpuagent/go.mod
	@touch ${TOP_DIR}/libamdsmi/go.mod
	@echo "setting up go mod packages"
	@go mod tidy
	@go mod edit -go=1.24.6
	#CVE-2024-24790 - amd-metrics-exporter
	@go mod edit -replace golang.org/x/net@v0.29.0=golang.org/x/net@v0.36.0
	#CVE-2025-30204 - amd-test-runner
	@go mod edit -replace github.com/golang-jwt/jwt/v5@v5.2.1=github.com/golang-jwt/jwt/v5@v5.2.2
	#CVE GHSA-fv92-fjc5-jj9h - amdgpuhealth
	@go mod edit -replace github.com/go-viper/mapstructure/v2@v2.2.1=github.com/go-viper/mapstructure/v2@v2.3.0
	@go mod vendor
	@rm ${TOP_DIR}/gpuagent/go.mod
	@rm ${TOP_DIR}/libamdsmi/go.mod

.PHONY: docs clean-docs dep-docs
dep-docs:
	pip install -r $(DOCS_DIR)/sphinx/requirements.txt

docs: dep-docs
	sphinx-build -b html $(DOCS_DIR) $(HTML_DIR)
	@echo "Docs built at $(HTML_DIR)/index.html"

clean-docs:
	rm -rf $(BUILD_DIR)


.PHONY: base-image
base-image:
	${MAKE} -C tools/base-image

copyrights:
	GOFLAGS=-mod=mod go run tools/build/copyright/main.go && ${MAKE} fmt && ./tools/build/check-local-files.sh

# target to update remote submodule repo for amdsmi and gpuagent
.PHONY: update-submodules
update-submodules:
	git submodule update --remote --recursive

.PHONY: e2e-test
e2e-test:
	$(MAKE) -C test/e2e

.PHONY: e2e
e2e:
	$(MAKE) docker-mock
	$(MAKE) e2e-test

.PHOHY: k8s-e2e
k8s-e2e:
	TOP_DIR=$(CURDIR) $(MAKE) -C test/k8s-e2e

.PHONY: helm-lint
helm-lint:
	#copy default config
	jq 'del(.ServerPort, .GPUConfig.ExtraPodLabels, .NICConfig)' $(CONFIG_DIR)/config.json > $(HELM_CHARTS_DIR)/config.json
	cd $(HELM_CHARTS_DIR); helm lint

.PHONY: helm-build
helm-build: helm-lint
	rm -rf helm-charts/device-metrics-exporter-charts*
	# updating project version in helm Chart.yaml
	sed -i -e 's|appVersion:.*$$|appVersion: "${REL_IMAGE_TAG}"|' helm-charts/Chart.yaml
	sed -i '0,/version:/s|version:.*|version: ${REL_IMAGE_TAG}|' helm-charts/Chart.yaml
	${MAKE} helm-docs
	helm package helm-charts/ --destination ./helm-charts --app-version ${REL_IMAGE_TAG} --version ${REL_IMAGE_TAG}
	cp -vf helm-charts/device-metrics-exporter-charts* helm-charts/device-metrics-exporter-charts.tgz

# cicd target to build helm chart - requires PROJECT_VERSION, EXPORTER_IMAGE_TAG to be set
.PHONY: helm
helm: helm-lint
	@rm -rf helm-charts-k8s
	@yq eval -i '.appVersion = "$(EXPORTER_IMAGE_TAG)"' helm-charts/Chart.yaml
	@yq eval -i '.version = "$(EXPORTER_IMAGE_TAG)"' helm-charts/Chart.yaml
	@yq eval -i '.image.repository = "$(HELM_EXPORTER_IMAGE)"' helm-charts/values.yaml
	@yq eval -i '.image.tag = "$(EXPORTER_IMAGE_TAG)"' helm-charts/values.yaml
	@mkdir -p helm-charts-k8s
	${MAKE} helm-docs
	helm package helm-charts/ --destination ./helm-charts-k8s --app-version ${PROJECT_VERSION} --version ${PROJECT_VERSION}
	cp -vf helm-charts-k8s/device-metrics-exporter-*.tgz helm-charts-k8s/device-metrics-exporter-helm-k8s-${PROJECT_VERSION}.tgz

.PHONY: slurm-sim
slurm-sim:
	${MAKE} -C pkg/exporter/scheduler/slurmsim TOP_DIR=$(CURDIR)

# create development build container only if there is changes done on
# tools/base-image/Dockerfile
.PHONY: build-dev-container
build-dev-container:
	${MAKE} -C tools/base-image all INSECURE_REGISTRY=$(INSECURE_REGISTRY)

.PHONY: amdsmi-compile-all
amdsmi-compile-all:
	${MAKE} amdsmi-compile-ub24
	${MAKE} amdsmi-compile-ub22
	${MAKE} amdsmi-compile-rhel
	#${MAKE} amdsmi-compile-azure

.PHONY: gimsmi-compile-all
gimsmi-compile-all:
	${MAKE} gimsmi-compile-ub24
	${MAKE} gimsmi-compile-ub22
	${MAKE} gimsmi-compile-rhel

# build all components
.PHONY: build-all
build-all: 
	${MAKE} amdsmi-compile-all
	${MAKE} gimsmi-compile-all
	# no need to run this everytime, we build and copy assets once
	#${MAKE} rocprofiler-compile
	#${MAKE} gpuagent-compile
	@echo "Docker image build is available under docker/ directory"
	${MAKE} docker

