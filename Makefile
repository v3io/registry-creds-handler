# Copyright 2022 Iguazio
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
VERSION ?= "unstable"
GOPATH ?= $(shell go env GOPATH)
OS_NAME = $(shell uname)
GIT_COMMIT ?= $(shell git rev-parse HEAD)
REGISTRY ?= "gcr.io/iguazio/"
REGISTRY_HANDLER_IMAGE_NAME ?= "$(REGISTRY)registry-creds-handler:$(VERSION)"

# Link flags
GO_LINK_FLAGS ?= -s -w
GO_LINK_FLAGS_INJECT_VERSION := $(GO_LINK_FLAGS) \
	-X github.com/v3io/version-go.gitCommit=$(GIT_COMMIT) \
	-X github.com/v3io/version-go.label=$(VERSION)

.PHONY: build
build:
	@echo "Building image"
	docker build \
		--build-arg GO_LINK_FLAGS="$(GO_LINK_FLAGS_INJECT_VERSION)" \
		--file cmd/registrycredshandler/Dockerfile \
		--tag $(REGISTRY_HANDLER_IMAGE_NAME) \
		.
	@echo "Done"

.PHONY: push
push:
	@echo "Pushing image"
	docker push $(REGISTRY_HANDLER_IMAGE_NAME)
	@echo "Done"

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: lint
lint:
	@echo Installing linters...
	@test -e $(GOPATH)/bin/impi || \
		(mkdir -p $(GOPATH)/bin && \
		curl -s https://api.github.com/repos/pavius/impi/releases/latest \
		| grep -i "browser_download_url.*impi.*$(OS_NAME)" \
		| cut -d : -f 2,3 \
		| tr -d \" \
		| wget -O $(GOPATH)/bin/impi -qi - \
		&& chmod +x $(GOPATH)/bin/impi)

	@test -e $(GOPATH)/bin/golangci-lint || \
	  	(curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.41.1)

	@echo Verifying imports...
	$(GOPATH)/bin/impi \
		--local github.com/v3io/registry-creds-handler/ \
		--scheme stdLocalThirdParty \
		./cmd/... ./pkg/...

	@echo Linting...
	$(GOPATH)/bin/golangci-lint run -v --timeout 3m0s
	@echo Done.

.PHONY: test-unit
test-unit:
	go test -v ./pkg/... -short
