# SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

TOOLS_DIR := hack/tools
include hack/tools.mk

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: check
check: format $(GOIMPORTS) $(GOLANGCI_LINT)
	@TOOLS_BIN_DIR="$(TOOLS_DIR)/bin" ./hack/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...
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
test: $(KUBEBUILDER_DIR) $(GINKGO)
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) ginkgo ${COVER_FLAG} -r cmd pkg plugin

.PHONY: generate
generate: $(VGOPATH)
	@VGOPATH=$(VGOPATH) ./hack/generate-code
	@go fmt ./cmd/... ./pkg/...

.PHONY: format
format:
	@go fmt ./cmd/... ./pkg/...