# SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: revendor
revendor:
	@GO111MODULE=on go mod vendor
	@GO111MODULE=on go mod tidy

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
test:
	GO111MODULE=on go test -mod=vendor ./pkg/... ./cmd/...

.PHONY: generate
generate:
	@./hack/generate-code
