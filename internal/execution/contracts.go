package execution

import (
	"encoding/json"
	"fmt"
	"maps"
	"math/big"

	_ "embed"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

//go:embed contracts.json
var bundledContractsJSON []byte

var systemContractAddresses = []string{
	appconfig.DefaultDepositContractAddress,
	"0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02",
	"0x0000F90827F1C53a10cb7A02335B175320002935",
	"0x0000BBdDc7CE488642fb579F8B00f3a590007251",
	"0x00000961Ef480Eb55e80D19ad83579A64c007002",
}

var utilsContractAddresses = []string{
	"0x4e59b44847b379578588920ca78fbf26c0b4956c",
	"0x1820a4B7618BdE71Dce8cdc73aAB6C95905faD24",
	"0x13b0D85CcB8bf860b6b79AF3029fCA081AE9beF2",
	"0xba5Ed099633D3B313e4D5F7bdc1305d3c28ba5Ed",
	"0xcA11bde05977b3631167028862bE2a173976CA11",
	"0x000000000022D473030F116dDEE9F6B43aC78BA3",
}

func addContractProfiles(alloc types.GenesisAlloc, cfg *appconfig.Config) error {
	registry, err := loadBundledContracts()
	if err != nil {
		return err
	}

	for _, profile := range cfg.Execution.Contracts {
		switch profile {
		case appconfig.ExecutionContractsProfileSystem:
			if err := addContracts(alloc, registry, systemContractAddresses, cfg); err != nil {
				return fmt.Errorf("execution.contracts profile %q: %w", profile, err)
			}
		case appconfig.ExecutionContractsProfileUtils:
			if err := addContracts(alloc, registry, utilsContractAddresses, cfg); err != nil {
				return fmt.Errorf("execution.contracts profile %q: %w", profile, err)
			}
		case appconfig.ExecutionContractsProfileAll:
			if err := addContracts(alloc, registry, systemContractAddresses, cfg); err != nil {
				return fmt.Errorf("execution.contracts profile %q: %w", profile, err)
			}
			if err := addContracts(alloc, registry, utilsContractAddresses, cfg); err != nil {
				return fmt.Errorf("execution.contracts profile %q: %w", profile, err)
			}
		case appconfig.ExecutionContractsProfileNone:
			continue
		default:
			return fmt.Errorf("execution.contracts has unsupported profile %q", profile)
		}
	}

	return nil
}

func addContracts(alloc types.GenesisAlloc, registry types.GenesisAlloc, sourceAddresses []string, cfg *appconfig.Config) error {
	for _, source := range sourceAddresses {
		target := source
		if common.HexToAddress(source) == common.HexToAddress(appconfig.DefaultDepositContractAddress) {
			target = cfg.Consensus.DepositContractAddress
		}
		if err := addContract(alloc, registry, source, target); err != nil {
			return err
		}
	}
	return nil
}

func addContract(alloc types.GenesisAlloc, registry types.GenesisAlloc, source string, target string) error {
	sourceAddress := common.HexToAddress(source)
	account, ok := registry[sourceAddress]
	if !ok {
		return fmt.Errorf("bundled contract %s is missing from contracts.json", source)
	}

	targetAddress := common.HexToAddress(target)
	if _, exists := alloc[targetAddress]; exists {
		return fmt.Errorf("bundled contract address %s conflicts with another selected contract", target)
	}
	alloc[targetAddress] = cloneAccount(account)
	return nil
}

func loadBundledContracts() (types.GenesisAlloc, error) {
	var alloc types.GenesisAlloc
	if err := json.Unmarshal(bundledContractsJSON, &alloc); err != nil {
		return nil, fmt.Errorf("parse bundled contracts.json: %w", err)
	}
	return alloc, nil
}

func cloneAccount(account types.Account) types.Account {
	clone := types.Account{
		Nonce: account.Nonce,
	}
	if account.Balance != nil {
		clone.Balance = new(big.Int).Set(account.Balance)
	}
	if account.Code != nil {
		clone.Code = append([]byte(nil), account.Code...)
	}
	if account.Storage != nil {
		clone.Storage = make(map[common.Hash]common.Hash, len(account.Storage))
		maps.Copy(clone.Storage, account.Storage)
	}
	return clone
}
