# AGENTS.md

This file applies to the whole repository.

## Project

`eth-genesis-generator` is a Go CLI that generates local Ethereum PoS devnet
genesis artifacts:

- execution-layer `genesis.json`
- consensus-layer `genesis.ssz`
- consensus `config.yaml`
- validator `mnemonics.yaml`
- validator keystores and keystore password
- `manifest.json`

The default target is Fulu on the consensus layer and Osaka active at execution
genesis.

## Layout

- `cmd/eth-genesis-generator`: `urfave/cli/v2` CLI entrypoint and command parsing.
- `internal/config`: user config schema, init template rendering, defaults,
  validation, fork version logic.
- `internal/execution`: go-ethereum execution genesis construction.
- `internal/consensus`: consensus config rendering and beacon genesis generation.
- `internal/artifacts`: output directory layout, file writing, manifest hashes.
- `internal/keystores`: validator keystore generation.

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

`init` only writes `<out>/genesis.yaml`; it must not generate execution,
consensus, validator, keystore, or manifest artifacts. It supports configurable
template flags such as `--network-name`, `--chain-id`, `--fork`, repeated
`--prefund ADDRESS=AMOUNT`, and `--force` for overwriting an existing config.

Validate the execution genesis with the module-pinned geth tool:

```bash
geth_datadir="$(mktemp -d)"
go tool github.com/ethereum/go-ethereum/cmd/geth init --datadir "$geth_datadir" ./artifacts/execution/genesis.json
```

## Engineering Notes

- Keep the CLI thin; put behavior in `internal/*` packages so tests can call it.
- Keep `generate --config ... --out ...` behavior stable when changing `init`.
- Validate init inputs before writing files, while allowing `genesis_time: 0`
  in the template.
- Use `github.com/ethereum/go-ethereum` types for execution genesis instead of
  hand-rolling JSON.
- Use `github.com/ethpandaops/eth-beacon-genesis` for beacon state generation
  instead of implementing SSZ or beacon-state structs locally.
- Preserve deterministic tests by fixing genesis time and mnemonic in tests.
- Run `gofmt -w -s . && go fix ./...` on changed Go files and `go mod tidy` after dependency changes.
- Treat generated `validators/mnemonics.yaml` as secret material. Do not commit
  generated artifacts unless a task explicitly asks for fixtures.
