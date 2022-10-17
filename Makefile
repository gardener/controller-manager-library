# SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

TOOLS_DIR := hack/tools
include hack/tools.mk

.PHONY: revendor
revendor:
	go mod vendor
	go mod tidy

.PHONY: check
check:
	@.ci/check

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build \
        -mod=vendor \
        -a -v \
        ./pkg/... ./cmd/...

.PHONY: build-local
build-local:
	@CGO_ENABLED=0 GO111MODULE=on go build \
	    -mod=vendor \
        ./pkg/... ./cmd/...

.PHONY: test
test: $(KUBEBUILDER_DIR)
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) go test -mod=vendor ./pkg/... ./cmd/...

.PHONY: generate
generate:
	@./hack/generate-code
