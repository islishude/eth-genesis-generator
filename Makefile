APP := eth-genesis-generator
CMD := ./cmd/eth-genesis-generator
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP)
GO_PACKAGES := ./...

.PHONY: all help build test lint format tidy check smoke clean

help:
	@printf '%s\n' \
		'Targets:' \
		'  build   Build the CLI into bin/eth-genesis-generator' \
		'  test    Run Go tests' \
		'  lint    Run go vet' \
		'  format  Run gofmt on cmd and internal packages' \
		'  check   Run format, tidy check' \
		'  smoke   Run init+generate in a temporary directory' \
		'  clean   Remove build output'

all: build lint format test

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD)

install:
	go install -trimpath -ldflags="-s -w" $(CMD)

test:
	go test $(GO_PACKAGES)

lint:
	go vet $(GO_PACKAGES)
	golangci-lint run --timeout 10m

format:
	gofmt -w -s .
	go fix ./...
	go mod tidy

check:
	gofmt -d .
	go fix -diff ./...
	go mod tidy -diff 

smoke:
	@tmpdir=$$(mktemp -d); \
	echo "using $$tmpdir"; \
	go run $(CMD) init --out "$$tmpdir/devnet"; \
	go run $(CMD) generate --config "$$tmpdir/devnet/genesis.yaml" --out "$$tmpdir/artifacts"; \
	test -s "$$tmpdir/artifacts/execution/genesis.json"; \
	test -s "$$tmpdir/artifacts/consensus/genesis.ssz"; \
	test -s "$$tmpdir/artifacts/manifest.json"

clean:
	rm -rf $(BIN_DIR)
