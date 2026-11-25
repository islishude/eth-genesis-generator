# AGENTS.md

These instructions apply throughout the repository. `ethgen` generates a
matched execution and consensus genesis for local Ethereum PoS devnets.
See [README.md](README.md) for CLI examples, configuration keys, and the
artifact layout.

## CLI contracts

Keep argument parsing and command wiring in `cmd/ethgen`. Put configuration
defaults and validation in `internal/config` and artifact orchestration in
`internal/artifacts`.

- `init` validates inputs before writing and writes only `<out>/genesis.yaml`.
  Accept `genesis_time: 0` in the template. Require `--force` to replace an
  existing template.
- `generate` reads the supplied `--config` and writes artifacts to `--out`.
  Preserve this workflow when changing template flags or `init`.
- `all` writes the config and generates artifacts in a new or empty output
  directory. Keep its template flags aligned with `init`.

## Genesis and validator invariants

- Build execution genesis with go-ethereum types in `internal/execution`.
  Generate the beacon state and SSZ through `eth-beacon-genesis` in
  `internal/consensus`.
- Keep Fulu and Osaka as the defaults. Selecting Gloas explicitly activates
  Gloas and Amsterdam at genesis. For fork changes, verify that the decoded
  beacon state's `eth1_data.block_hash` matches the execution genesis hash.
  Check `latest_block_hash` as well for Gloas.
- For bundled predeploy changes, update
  [contracts.json](internal/execution/contracts.json) and
  [contracts.md](internal/execution/contracts.md) together. Verify which
  profiles and forks include the affected accounts.
- Use a fixed genesis time and test mnemonic in deterministic tests. Treat
  `validators/` and any config containing a mnemonic as private. Keep
  generated artifacts out of commits unless the task explicitly calls for
  fixtures.

## Verification

- For Go changes, run `gofmt -w -s . && go fix ./...`. Run `go mod tidy`
  after dependency changes. The Go CI checks are `make build`,
  `make check`, `make lint`, and `make test`.
- For changes to CLI or artifact generation, run `make smoke` to exercise the
  default `init` and `generate` workflow in a temporary directory.
- For fork changes, generate and decode beacon genesis SSZ with both
  `mainnet` and `minimal` presets. For execution genesis changes, import
  the generated file into a fresh data directory with the module-pinned
  Geth using the command in [Verify and develop](README.md#verify-and-develop).
