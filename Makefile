VERSION ?= "unstable"
GIT_COMMIT ?= $(shell git rev-parse HEAD)
REGISTRY_HANDLER_IMAGE_NAME ?= "gcr.io/iguazio/registry-creds-handler:$(VERSION)"

# Link flags
GO_LINK_FLAGS ?= -s -w
GO_LINK_FLAGS_INJECT_VERSION := $(GO_LINK_FLAGS) \
	-X github.com/v3io/version-go.gitCommit=$(GIT_COMMIT) \
	-X github.com/v3io/version-go.label=$(VERSION)

.PHONY: build
build:
	docker build \
		--build-arg GO_LINK_FLAGS="$(GO_LINK_FLAGS_INJECT_VERSION)" \
		--file cmd/registrycredshandler/Dockerfile \
		--tag $(REGISTRY_HANDLER_IMAGE_NAME) \
		.

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
		--skip pkg/platform/kube/apis \
		--skip pkg/platform/kube/client \
		./cmd/... ./pkg/...

	@echo Linting...
	$(GOPATH)/bin/golangci-lint run -v
	@echo Done.

.PHONY: test
test:
	go test -v ./pkg/...