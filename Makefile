# SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

#########################################
# Tools                                 #
#########################################

TOOLS_BIN_DIR ?= hack/tools/bin/$(go env GOOS)-$(go env GOARCH)
TOOLS_DIR := hack/tools
include hack/tools.mk

.PHONY: tidy
tidy:
	@go work use
	@go work sync
	@go mod tidy
	@cd pkg/internal/tools; go mod tidy

.PHONY: check
check: sast-report fastcheck

.PHONY: fastcheck
fastcheck: format $(GOIMPORTS) $(GOLANGCI_LINT)
	@TOOLS_BIN_DIR="$(TOOLS_BIN_DIR)" ./hack/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...
	@echo "Running go vet..."
	@go vet ./cmd/... ./pkg/...

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -a -v \
        ./pkg/... ./cmd/...

.PHONY: build-local
build-local:
	@CGO_ENABLED=0 go build \
        ./pkg/... ./cmd/...

.PHONY: test
test: $(SETUP_ENVTEST) $(GINKGO)
	KUBEBUILDER_ASSETS="$(shell realpath $(shell $(SETUP_ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(KUBEBUILDER_DIR) -p path))" ginkgo ${COVER_FLAG} -r cmd pkg plugin

.PHONY: generate
generate:
	./hack/generate-code
	@go fmt ./cmd/... ./pkg/...

.PHONY: format
format:
	@go fmt ./cmd/... ./pkg/...

.PHONY: sast
sast: $(GOSEC)
	@./hack/sast.sh

.PHONY: sast-report
sast-report: $(GOSEC)
	@./hack/sast.sh --gosec-report true