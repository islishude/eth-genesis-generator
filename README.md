# eth-genesis-generator

Generate matching execution and consensus genesis files for a local Ethereum
proof-of-stake devnet, along with the keys for its genesis validators. The
default configuration starts with Osaka on the execution layer and Fulu on the
consensus layer, chain ID `32382`, and 64 validators with 32 ETH each.

## Quick start

From the repository root, build with Go 1.27.4 or later, Make, and a C compiler.

```sh
make build
./bin/eth-genesis-generator all --out ./devnet
```

`all` writes the input configuration and every generated artifact to a new or
empty directory. Its default output directory is `./devnet`. With the default
configuration, generation sets genesis time to the current Unix time plus 60
seconds and creates a fresh 24-word validator mnemonic. The resolved timestamp
is recorded in `manifest.json`, and the mnemonic is saved in
`validators/mnemonics.yaml`.

To run a devnet with the generated files, see the
[client setup guide](example/example.md) for Reth, Geth, Lighthouse, and Prysm.
The guide also includes a Docker Compose setup that builds the generator and
starts a local Reth and Lighthouse network.

## Customize the genesis

Use `init` to write a template, edit `genesis.yaml`, then pass it to
`generate`.

```sh
./bin/eth-genesis-generator init --out ./custom-devnet --chain-id 12345 \
  --prefund 0x1000000000000000000000000000000000000001=1000000000000000000
./bin/eth-genesis-generator generate \
  --config ./custom-devnet/genesis.yaml --out ./artifacts
```

`init` writes only `genesis.yaml` and accepts `--force` to replace an existing
template. `generate` requires `--config` and defaults to `./artifacts`. Use a
fresh output directory for each network to keep its files together.

Both `init` and `all` accept the configuration flags shown by `--help`.
`--prefund ADDRESS=AMOUNT` is repeatable and takes a decimal amount in wei.
For YAML, use a quoted decimal string for each amount in `execution.prefund`.

| Config key | Purpose |
| --- | --- |
| `network.chain_id` | Sets the execution chain ID and consensus deposit chain and network IDs. |
| `network.genesis_time` | Sets the shared Unix genesis timestamp in seconds. Leave it at `0` to resolve it during generation. |
| `execution.contracts` | Selects bundled contract predeploys. The default list is empty. |
| `consensus.fork` | Selects the consensus fork active at genesis. The default is `fulu`. |
| `consensus.preset_base` | Selects the consensus preset. The default is `mainnet`. Set `minimal` for clients configured for the minimal preset. |
| `consensus.validator_count` | Sets the number of validators included in the beacon genesis state. |
| `consensus.mnemonic` | Supplies a BIP-39 mnemonic for validator key derivation. Leave it empty to generate a new one. |
| `consensus.output_json` | Controls whether generation also writes `consensus/genesis.json`. Defaults to `true`. |

To regenerate the same chain genesis and validator keys, reuse the resolved
genesis timestamp and mnemonic with the other settings from the original run.
Each run encrypts the keystores with a fresh password.

### Glamsterdam

Select `--fork gloas` with `init` or `all`, or set `consensus.fork: gloas` in
`genesis.yaml`. This activates Gloas on the consensus layer and Amsterdam on
the execution layer at genesis. Use client builds that support those forks.

```sh
./bin/eth-genesis-generator all --out ./glamsterdam --fork gloas \
  --execution-contracts system
```

### Contract predeploys

Genesis validators are included directly in the beacon state. To also predeploy
the validator deposit contract and other system contracts on the execution
layer, choose `system`.

| Profile | Contracts |
| --- | --- |
| `system` | System contracts, including the validator deposit contract at `consensus.deposit_contract_address`. With Gloas, this also includes the EIP-8282 builder deposit and exit contracts. |
| `utils` | Utility contracts such as Multicall3 and Permit2. |
| `all` | Both groups. |
| `none` | Empty set of predeploys. |

Pass profiles as a comma-separated flag such as `--execution-contracts system,utils`,
or use a YAML list such as `contracts: [system, utils]` under `execution`.
Use `all` and `none` individually. See the
[bundled contract list](internal/execution/contracts.md) for addresses.

## Generated files

| Path | Use |
| --- | --- |
| `execution/genesis.json` | Initialize the execution client's chain database. |
| `consensus/config.yaml` | Configure the beacon node and validator client for the same network. |
| `consensus/genesis.ssz` | Initialize the beacon node with the genesis state. |
| `consensus/genesis.json` | Inspect the beacon genesis state as JSON. |
| `validators/mnemonics.yaml` | Record the mnemonic and derivation range used for the genesis validators. |
| `validators/keystores/*.json` | Import the EIP-2335 validator keys into a validator client. |
| `validators/keystores/password.txt` | Decrypt the generated keystores during import. |
| `manifest.json` | Check the chain ID, resolved genesis time, execution genesis block hash, beacon state version, validator counts, and SHA-256 digests of generated files. |

Keep `validators/` and any `genesis.yaml` containing a mnemonic private.

## Verify and develop

From the repository root, import the execution genesis into a temporary
database with the version of Geth pinned in `go.mod`. This example uses the
output from the quick start.

```sh
geth_datadir="$(mktemp -d)"
go tool github.com/ethereum/go-ethereum/cmd/geth init \
  --datadir "$geth_datadir" ./devnet/execution/genesis.json
```

Run `make test` for the Go test suite and `make smoke` to exercise
`init` and `generate` in a temporary directory. The `make lint` and
`make check` targets require `golangci-lint`.

[MIT License](LICENSE)
