# SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

SYSTEM_NAME              := $(shell uname -s | tr '[:upper:]' '[:lower:]')
SYSTEM_ARCH              := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
TOOLS_BIN_DIR            := $(TOOLS_DIR)/bin/$(SYSTEM_NAME)-$(SYSTEM_ARCH)
CONTROLLER_GEN           := $(TOOLS_BIN_DIR)/controller-gen
GOLANGCI_LINT            := $(TOOLS_BIN_DIR)/golangci-lint
GOSEC                    := $(TOOLS_BIN_DIR)/gosec
GOIMPORTS                := $(TOOLS_BIN_DIR)/goimports
GINKGO                   := $(TOOLS_BIN_DIR)/ginkgo

SETUP_ENVTEST            := $(TOOLS_BIN_DIR)/setup-envtest
ENVTEST_K8S_VERSION      := 1.35.0
CONTROLLER_RUNTIME_VERSION ?= $(call version_gomod,sigs.k8s.io/controller-runtime)
KUBEBUILDER_DIR          := $(TOOLS_BIN_DIR)/kubebuilder

export TOOLS_BIN_DIR := $(TOOLS_BIN_DIR)
export PATH := $(abspath $(TOOLS_BIN_DIR)):$(PATH)

# Use this function to get the version of a go module from go.mod
version_gomod = $(shell go list $(MODFILE_TOOL_MOD) -f '{{ .Version }}' -m $(1))

# Use this function to copy the tool binary built by Go to the location passed as arg.
#   E.g., `$(call go_tool_copy,./path/to/tool)` will copy the tool binary built by Go to ./path/to/tool.
go_tool_copy = $(shell cp $$(go tool $(MODFILE_TOOL_MOD) -n $$(basename $(1))) $(1))

# tool versions from go.mod
GOLANGCI_LINT_VERSION   ?= $(call version_gomod,github.com/golangci/golangci-lint/v2)
GOSEC_VERSION           ?= $(call version_gomod,github.com/securego/gosec/v2)
CONTROLLER_GEN_VERSION 	?= $(call version_gomod,sigs.k8s.io/controller-tools)
GINKGO_VERSION          ?= $(call version_gomod,github.com/onsi/ginkgo/v2)
GOIMPORTS_VERSION       ?= $(call version_gomod,golang.org/x/tools)

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
	$(call go_tool_copy,$(CONTROLLER_GEN))

$(GOLANGCI_LINT): $(call tool_version_file,$(GOLANGCI_LINT),$(GOLANGCI_LINT_VERSION))
	$(call go_tool_copy,$(GOLANGCI_LINT))

$(GOIMPORTS): $(call tool_version_file,$(GOIMPORTS),$(GOIMPORTS_VERSION))
	$(call go_tool_copy,$(GOIMPORTS))

$(GINKGO): $(call tool_version_file,$(GINKGO),$(GINKGO_VERSION))
	$(call go_tool_copy,$(GINKGO))

$(SETUP_ENVTEST): $(call tool_version_file,$(SETUP_ENVTEST),$(CONTROLLER_RUNTIME_VERSION))
	curl -Lo $(SETUP_ENVTEST) https://github.com/kubernetes-sigs/controller-runtime/releases/download/$(CONTROLLER_RUNTIME_VERSION)/setup-envtest-$(SYSTEM_NAME)-$(SYSTEM_ARCH)
	chmod +x $(SETUP_ENVTEST)

$(GOSEC): $(call tool_version_file,$(GOSEC),$(GOSEC_VERSION))
	$(call go_tool_copy,$(GOSEC))
