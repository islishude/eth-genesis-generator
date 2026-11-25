package artifacts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethpandaops/eth-beacon-genesis/beaconconfig"
	"github.com/ethpandaops/eth-beacon-genesis/beaconutils"
	"github.com/ethpandaops/go-eth2-client/spec/gloas"

	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

func TestGenerateArtifacts(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Network.GenesisTime = 1_700_000_100
	cfg.Consensus.ValidatorCount = 1
	cfg.Consensus.Mnemonic = "legal winner thank year wave sausage worth useful legal winner thank yellow"

	outDir := t.TempDir()
	manifest, err := Generate(&cfg, outDir, time.Unix(1_700_000_001, 0))
	if err != nil {
		t.Fatal(err)
	}

	if manifest.StateVersion != "fulu" {
		t.Fatalf("state version = %s", manifest.StateVersion)
	}
	if manifest.ValidatorCount != 1 {
		t.Fatalf("validator count = %d", manifest.ValidatorCount)
	}
	if manifest.GeneratedMnemonic {
		t.Fatal("mnemonic should not be marked generated")
	}

	for _, relPath := range []string{
		"execution/genesis.json",
		"consensus/config.yaml",
		"consensus/genesis.ssz",
		"consensus/genesis.json",
		"validators/mnemonics.yaml",
		"validators/keystores/password.txt",
		"validators/keystores/keystore-000000.json",
		"manifest.json",
	} {
		if _, err := os.Stat(filepath.Join(outDir, filepath.FromSlash(relPath))); err != nil {
			t.Fatalf("missing %s: %v", relPath, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ArtifactHashes["consensus/genesis.ssz"] == "" {
		t.Fatal("missing genesis.ssz artifact hash")
	}
	if decoded.ValidatorKeystoreCount != 1 {
		t.Fatalf("validator keystore count = %d", decoded.ValidatorKeystoreCount)
	}
	if decoded.ArtifactHashes["validators/keystores/password.txt"] == "" {
		t.Fatal("missing validator keystore password artifact hash")
	}
	if decoded.ArtifactHashes["validators/keystores/keystore-000000.json"] == "" {
		t.Fatal("missing validator keystore artifact hash")
	}
}

func TestGloasGenesisSSZ(t *testing.T) {
	for _, preset := range []string{"mainnet", "minimal"} {
		t.Run(preset, func(t *testing.T) {
			now := time.Unix(1_700_000_000, 0)
			var cfg appconfig.Config
			cfg.ApplyDefaults(now)
			cfg.Network.GenesisTime = 1_700_000_100
			cfg.Consensus.Fork = "gloas"
			cfg.Consensus.ForkVersions = map[string]string{"gloas": "0x07010203"}
			cfg.Consensus.PresetBase = preset
			cfg.Consensus.ValidatorCount = 1
			cfg.Consensus.Mnemonic = "legal winner thank year wave sausage worth useful legal winner thank yellow"
			if err := cfg.Validate(); err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			manifest, err := Generate(&cfg, dir, now)
			if err != nil {
				t.Fatal(err)
			}
			if manifest.StateVersion != "gloas" || manifest.ValidatorCount != 1 {
				t.Fatalf("manifest state version %s, validators %d", manifest.StateVersion, manifest.ValidatorCount)
			}
			spec, err := beaconconfig.LoadConfig(filepath.Join(dir, "consensus/config.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(dir, "consensus/genesis.ssz"))
			if err != nil {
				t.Fatal(err)
			}
			var state gloas.BeaconState
			if err := beaconutils.GetDynSSZ(spec).UnmarshalSSZ(&state, data); err != nil {
				t.Fatal(err)
			}
			if state.Fork.CurrentVersion != [4]byte{7, 1, 2, 3} || state.Fork.Epoch != 0 || len(state.Validators) != 1 {
				t.Fatalf("SSZ fork %v, validators %d", state.Fork, len(state.Validators))
			}
			hash := common.HexToHash(manifest.ExecutionGenesisHash)
			if common.BytesToHash(state.ETH1Data.BlockHash) != hash || common.Hash(state.LatestBlockHash) != hash {
				t.Fatalf("SSZ execution hashes differ from manifest hash %s", hash)
			}
		})
	}
}
