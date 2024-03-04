# SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

TOOLS_BIN_DIR            := $(TOOLS_DIR)/bin
KUBEBUILDER_K8S_VERSION  := 1.28.0
KUBEBUILDER_TAG          := $(TOOLS_BIN_DIR)/kubebuilder
KUBEBUILDER_DIR          := $(TOOLS_BIN_DIR)/kubebuilder_$(KUBEBUILDER_K8S_VERSION)
KUBEBUILDER_ASSETS       := "$(shell realpath $(KUBEBUILDER_DIR))/bin"
CONTROLLER_GEN           := $(TOOLS_BIN_DIR)/controller-gen
GOLANGCI_LINT            := $(TOOLS_BIN_DIR)/golangci-lint
GOIMPORTS                := $(TOOLS_BIN_DIR)/goimports
GINKGO                   := $(TOOLS_BIN_DIR)/ginkgo
VGOPATH                  := $(TOOLS_BIN_DIR)/vgopath

export TOOLS_BIN_DIR := $(TOOLS_BIN_DIR)
export PATH := $(abspath $(TOOLS_BIN_DIR)):$(PATH)

GOLANGCI_LINT_VERSION ?= v1.55.1
VGOPATH_VERSION ?= v0.1.3

# Use this function to get the version of a go module from go.mod
version_gomod = $(shell go list -mod=mod -f '{{ .Version }}' -m $(1))

# tool versions from go.mod
CONTROLLER_GEN_VERSION ?= $(call version_gomod,sigs.k8s.io/controller-tools)
GINKGO_VERSION ?= $(call version_gomod,github.com/onsi/ginkgo/v2)
GOIMPORTS_VERSION ?= $(call version_gomod,golang.org/x/tools)

# Use this "function" to add the version file as a prerequisite for the tool target: e.g.
#   $(HELM): $(call tool_version_file,$(HELM),$(HELM_VERSION))
tool_version_file = $(TOOLS_BIN_DIR)/.version_$(subst $(TOOLS_BIN_DIR)/,,$(1))_$(2)

# This target cleans up any previous version files for the given tool and creates the given version file.
# This way, we can generically determine, which version was installed without calling each and every binary explicitly.
$(TOOLS_BIN_DIR)/.version_%:
	@mkdir -p $(TOOLS_BIN_DIR)
	@version_file=$@; rm -f $${version_file%_*}*
	@touch $@

$(CONTROLLER_GEN): $(call tool_version_file,$(CONTROLLER_GEN),$(CONTROLLER_GEN_VERSION))
	go build -o $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen

$(GOLANGCI_LINT): $(call tool_version_file,$(GOLANGCI_LINT),$(GOLANGCI_LINT_VERSION))
	@# CGO_ENABLED has to be set to 1 in order for golangci-lint to be able to load plugins
	@# see https://github.com/golangci/golangci-lint/issues/1276
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) CGO_ENABLED=1 go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(GOIMPORTS): $(call tool_version_file,$(GOIMPORTS),$(GOIMPORTS_VERSION))
	go build -o $(GOIMPORTS) golang.org/x/tools/cmd/goimports

$(GINKGO): $(call tool_version_file,$(GINKGO),$(GINKGO_VERSION))
	go build -o $(GINKGO) github.com/onsi/ginkgo/v2/ginkgo

$(KUBEBUILDER_DIR): $(call tool_version_file,$(KUBEBUILDER_TAG),$(KUBEBUILDER_K8S_VERSION))
	curl -sSL https://go.kubebuilder.io/test-tools/$(KUBEBUILDER_K8S_VERSION)/$(shell go env GOOS)/$(shell go env GOARCH) | tar -xvz
	@mv kubebuilder $(KUBEBUILDER_DIR)
	@touch $(KUBEBUILDER_TAG)

$(VGOPATH): $(call tool_version_file,$(VGOPATH),$(VGOPATH_VERSION))
	go build -o $(VGOPATH) github.com/ironcore-dev/vgopath
