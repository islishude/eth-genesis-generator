# AGENTS.md

This file applies to the whole repository.

## Project

`eth-genesis-generator` is a Go CLI that generates local Ethereum PoS devnet
genesis artifacts:

- execution-layer `genesis.json`
- consensus-layer `genesis.ssz`
- consensus `config.yaml`
- validator `mnemonics.yaml`
- `manifest.json`

The default target is Fulu on the consensus layer and Osaka active at execution
genesis. v1 uses pre-filled validators and does not predeploy deposit contract
bytecode.

## Layout

- `cmd/eth-genesis-generator`: CLI entrypoint and command parsing.
- `internal/config`: user config schema, defaults, validation, fork version logic.
- `internal/execution`: go-ethereum execution genesis construction.
- `internal/consensus`: consensus config rendering and beacon genesis generation.
- `internal/artifacts`: output directory layout, file writing, manifest hashes.

## Common Commands

```bash
make build
make test
make lint
make format
make smoke
go run ./cmd/eth-genesis-generator init --out ./devnet
go run ./cmd/eth-genesis-generator generate --config ./devnet/genesis.yaml --out ./artifacts
```

Validate the execution genesis with the module-pinned geth tool:

```bash
geth_datadir="$(mktemp -d)"
go tool github.com/ethereum/go-ethereum/cmd/geth init --datadir "$geth_datadir" ./artifacts/execution/genesis.json
```

## Engineering Notes

- Keep the CLI thin; put behavior in `internal/*` packages so tests can call it.
- Use `github.com/ethereum/go-ethereum` types for execution genesis instead of
  hand-rolling JSON.
- Use `github.com/ethpandaops/eth-beacon-genesis` for beacon state generation
  instead of implementing SSZ or beacon-state structs locally.
- Preserve deterministic tests by fixing genesis time and mnemonic in tests.
- Run `gofmt` on changed Go files and `go mod tidy` after dependency changes.
- Treat generated `validators/mnemonics.yaml` as secret material. Do not commit
  generated artifacts unless a task explicitly asks for fixtures.
