package execution

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

func TestBuildGenesisIsOsakaAtGenesis(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Network.GenesisTime = 1_700_000_100
	cfg.Execution.Prefund["0x1000000000000000000000000000000000000001"] = "1000"

	genesis, err := BuildGenesis(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if genesis.Config.OsakaTime == nil || *genesis.Config.OsakaTime != 0 {
		t.Fatalf("osaka time = %v", genesis.Config.OsakaTime)
	}
	block := genesis.ToBlock()
	if block.BlobGasUsed() == nil || block.ExcessBlobGas() == nil {
		t.Fatal("expected blob gas fields on genesis block")
	}
	if block.BaseFee() == nil {
		t.Fatal("expected base fee")
	}

	data, err := MarshalGenesis(genesis)
	if err != nil {
		t.Fatal(err)
	}
	var decoded core.Genesis
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Config.ChainID.Uint64() != cfg.Network.ChainID {
		t.Fatalf("decoded chain id = %d", decoded.Config.ChainID.Uint64())
	}
}

func TestBuildGenesisAddsSystemContracts(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Execution.Contracts = []string{appconfig.ExecutionContractsProfileSystem}

	genesis, err := BuildGenesis(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	deposit := genesis.Alloc[common.HexToAddress(appconfig.DefaultDepositContractAddress)]
	if len(deposit.Code) == 0 {
		t.Fatal("expected deposit contract code")
	}
	if len(deposit.Storage) == 0 {
		t.Fatal("expected deposit contract storage")
	}
	if deposit.Nonce != 1 {
		t.Fatalf("deposit nonce = %d", deposit.Nonce)
	}
	if deposit.Balance == nil || deposit.Balance.Sign() != 0 {
		t.Fatalf("deposit balance = %v", deposit.Balance)
	}

	beaconRoots := genesis.Alloc[common.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02")]
	if len(beaconRoots.Code) == 0 || beaconRoots.Nonce != 1 {
		t.Fatalf("beacon roots account = %#v", beaconRoots)
	}
	if _, ok := genesis.Alloc[common.HexToAddress("0x4e59b44847b379578588920ca78fbf26c0b4956c")]; ok {
		t.Fatal("utils contract should not be included by system profile")
	}
}

func TestBuildGenesisAddsUtilsContracts(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Execution.Contracts = []string{appconfig.ExecutionContractsProfileUtils}

	genesis, err := BuildGenesis(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	deployer := genesis.Alloc[common.HexToAddress("0x4e59b44847b379578588920ca78fbf26c0b4956c")]
	if len(deployer.Code) == 0 || deployer.Nonce != 1 {
		t.Fatalf("deterministic create2 deployer account = %#v", deployer)
	}
	if _, ok := genesis.Alloc[common.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02")]; ok {
		t.Fatal("system contract should not be included by utils profile")
	}
}

func TestBuildGenesisPrefundPreservesContractAccountData(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Execution.Contracts = []string{appconfig.ExecutionContractsProfileSystem}
	cfg.Execution.Prefund[appconfig.DefaultDepositContractAddress] = "123"

	genesis, err := BuildGenesis(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	deposit := genesis.Alloc[common.HexToAddress(appconfig.DefaultDepositContractAddress)]
	if deposit.Balance.Cmp(big.NewInt(123)) != 0 {
		t.Fatalf("deposit balance = %s", deposit.Balance)
	}
	if len(deposit.Code) == 0 {
		t.Fatal("expected prefund to preserve deposit contract code")
	}
	if len(deposit.Storage) == 0 {
		t.Fatal("expected prefund to preserve deposit contract storage")
	}
	if deposit.Nonce != 1 {
		t.Fatalf("deposit nonce = %d", deposit.Nonce)
	}
}

func TestBuildGenesisMapsDepositContractToConfiguredAddress(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Execution.Contracts = []string{appconfig.ExecutionContractsProfileSystem}
	cfg.Consensus.DepositContractAddress = "0x1000000000000000000000000000000000000005"

	genesis, err := BuildGenesis(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	target := common.HexToAddress(cfg.Consensus.DepositContractAddress)
	if genesis.Config.DepositContractAddress != target {
		t.Fatalf("chain config deposit address = %s", genesis.Config.DepositContractAddress)
	}
	deposit := genesis.Alloc[target]
	if len(deposit.Code) == 0 || len(deposit.Storage) == 0 {
		t.Fatalf("configured deposit account = %#v", deposit)
	}
	if _, ok := genesis.Alloc[common.HexToAddress(appconfig.DefaultDepositContractAddress)]; ok {
		t.Fatal("default deposit address should not be allocated when deposit address is customized")
	}
}
