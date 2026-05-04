APP := eth-genesis-generator
CMD := ./cmd/eth-genesis-generator
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP)
GO_PACKAGES := ./...

.PHONY: help build test lint format tidy check smoke clean

help:
	@printf '%s\n' \
		'Targets:' \
		'  build   Build the CLI into bin/eth-genesis-generator' \
		'  test    Run Go tests' \
		'  lint    Run go vet' \
		'  format  Run gofmt on cmd and internal packages' \
		'  tidy    Run go mod tidy' \
		'  check   Run format, tidy, lint, and test' \
		'  smoke   Run init+generate in a temporary directory' \
		'  clean   Remove build output'

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

check:
	gofmt -d .
	go fix -diff ./...
	go mod tidy -diff 

tidy:
	go mod tidy

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
