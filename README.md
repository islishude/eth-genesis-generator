# eth-genesis-generator

Go CLI for generating local Ethereum proof-of-stake devnet genesis artifacts.

The default profile creates an execution-layer genesis with Osaka active at
genesis and a consensus-layer Fulu `genesis.ssz` with pre-filled validators.
It does not predeploy the deposit contract bytecode in v1.

## Usage

```bash
eth-genesis-generator all --out ./devnet
```

`all` writes `genesis.yaml`, `manifest.json`, and the execution, consensus, and
validator artifacts together. The output directory defaults to `./devnet` and
must be new or empty. Use the same configuration flags as `init`, for example
`--fork gloas --chain-id 12345`. With the default settings, the template keeps
`genesis_time: 0` and an empty mnemonic. Generation records the resolved time
(current time plus 60 seconds) in the manifest and writes the generated mnemonic
to `validators/mnemonics.yaml`.

To edit the configuration before generating artifacts:

```bash
eth-genesis-generator init --out ./devnet
# Edit ./devnet/genesis.yaml, then run:
eth-genesis-generator generate --config ./devnet/genesis.yaml --out ./artifacts
```

To activate Glamsterdam at genesis (Gloas on consensus and Amsterdam on execution),
select `--fork gloas` or set `consensus.fork: gloas` in `genesis.yaml`:

```bash
eth-genesis-generator init --fork gloas --out ./glamsterdam
eth-genesis-generator generate --config ./glamsterdam/genesis.yaml --out ./artifacts
```

For Gloas, `init --execution-contracts system` also includes the EIP-8282
builder deposit and exit contracts. The `all` profile includes them as well.

By default `init` writes `execution.contracts: []`, so generated execution
genesis files do not predeploy bundled contracts. To include optional predeploys
from [internal/execution/contracts.md](./internal/execution/contracts.md), write
one of the supported profiles into `genesis.yaml`:

```bash
eth-genesis-generator init --out ./devnet --execution-contracts system
eth-genesis-generator init --out ./devnet --execution-contracts system,utils
eth-genesis-generator init --out ./devnet --execution-contracts all
```

See [example.md](./example/example.md) for local Geth/Reth and Prysm/Lighthouse
integration examples.

## Development

```bash
make build
make test
make lint
make format
make smoke
```

Generated files:

- `execution/genesis.json`
- `consensus/config.yaml`
- `consensus/genesis.ssz`
- `consensus/genesis.json`
- `validators/mnemonics.yaml`
- `validators/keystores/*.json`
- `validators/keystores/password.txt`
- `manifest.json`

`validators/mnemonics.yaml` and `validators/keystores/` contain validator
secret material. Treat them as secrets for any network that carries value.

To validate the execution genesis without a separately installed `geth` binary:

```bash
geth_datadir="$(mktemp -d)"
go tool github.com/ethereum/go-ethereum/cmd/geth init --datadir "$geth_datadir" ./artifacts/execution/genesis.json
```
