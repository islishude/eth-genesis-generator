package execution

import (
	"encoding/json"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

// BuildGenesis constructs a geth core.Genesis with all execution forks active at genesis.
func BuildGenesis(cfg *appconfig.Config) (*core.Genesis, error) {
	zero := uint64(0)
	baseFee, err := appconfig.ParseBigInt(cfg.Execution.BaseFeePerGas)
	if err != nil {
		return nil, err
	}
	extraData, err := appconfig.DecodeHexBytes(cfg.Execution.ExtraData)
	if err != nil {
		return nil, err
	}

	alloc, err := buildAlloc(cfg.Execution.Prefund)
	if err != nil {
		return nil, err
	}

	return &core.Genesis{
		Config:     buildChainConfig(cfg),
		Nonce:      0,
		Timestamp:  cfg.Network.GenesisTime,
		ExtraData:  extraData,
		GasLimit:   cfg.Execution.GasLimit,
		Difficulty: big.NewInt(0),
		Mixhash:    common.Hash{},
		Coinbase:   common.Address{},
		Alloc:      alloc,
		BaseFee:    baseFee,
		// Osaka genesis blocks must carry blob gas fields so geth accepts the file.
		BlobGasUsed: func() *uint64 {
			v := zero
			return &v
		}(),
		ExcessBlobGas: func() *uint64 {
			v := zero
			return &v
		}(),
	}, nil
}

// MarshalGenesis renders a geth genesis in the canonical indented JSON form.
func MarshalGenesis(genesis *core.Genesis) ([]byte, error) {
	return json.MarshalIndent(genesis, "", "  ")
}

func buildAlloc(prefund map[string]string) (types.GenesisAlloc, error) {
	alloc := make(types.GenesisAlloc, len(prefund))
	addresses := make([]string, 0, len(prefund))
	for address := range prefund {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	for _, address := range addresses {
		balance, err := appconfig.ParseBigInt(prefund[address])
		if err != nil {
			return nil, err
		}
		alloc[common.HexToAddress(address)] = types.Account{Balance: balance}
	}
	return alloc, nil
}

func buildChainConfig(cfg *appconfig.Config) *params.ChainConfig {
	return &params.ChainConfig{
		ChainID: new(big.Int).SetUint64(cfg.Network.ChainID),
		// Historical execution forks activate at block 0 for a post-merge devnet.
		HomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          false,
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		GrayGlacierBlock:        big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		TerminalTotalDifficulty: big.NewInt(0),
		// Timestamp-based forks activate at genesis; Osaka is the default EL target.
		ShanghaiTime:           new(uint64(0)),
		CancunTime:             new(uint64(0)),
		PragueTime:             new(uint64(0)),
		OsakaTime:              new(uint64(0)),
		DepositContractAddress: common.HexToAddress(cfg.Consensus.DepositContractAddress),
		Ethash:                 new(params.EthashConfig),
		BlobScheduleConfig: &params.BlobScheduleConfig{
			Cancun: params.DefaultCancunBlobConfig,
			Prague: params.DefaultPragueBlobConfig,
			Osaka:  params.DefaultOsakaBlobConfig,
		},
	}
}
