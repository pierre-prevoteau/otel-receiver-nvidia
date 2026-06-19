# Makefile for the NVIDIA GPU OpenTelemetry receiver.

MDATAGEN_VERSION ?= v0.154.0
TOOLS_DIR := $(CURDIR)/.tools/mdatagen
METADATA := $(CURDIR)/nvidiareceiver/metadata.yaml

.PHONY: all
all: generate tidy test

# Regenerate internal/metadata, documentation.md and the generated tests from
# nvidiareceiver/metadata.yaml.
#
# mdatagen cannot be installed with `go install ...@version` because its module
# ships replace directives, so it is run from an isolated tool module (Go 1.24+).
.PHONY: generate
generate:
	@mkdir -p $(TOOLS_DIR)
	@cd $(TOOLS_DIR) && \
		(test -f go.mod || go mod init mdatagen-runner >/dev/null 2>&1) && \
		go get -tool go.opentelemetry.io/collector/cmd/mdatagen@$(MDATAGEN_VERSION) >/dev/null 2>&1 && \
		go tool mdatagen $(METADATA)

.PHONY: test
test:
	go test ./...

# Regenerate the scraper golden file from the current emitted metrics.
.PHONY: golden
golden:
	WRITE_GOLDEN=true go test ./nvidiareceiver/ -run TestScrape$$ || true

.PHONY: build
build:
	go build ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	gofmt -w $(CURDIR)/nvidiareceiver
