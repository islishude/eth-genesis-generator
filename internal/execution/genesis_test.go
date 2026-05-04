package execution

import (
	"encoding/json"
	"testing"
	"time"

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
