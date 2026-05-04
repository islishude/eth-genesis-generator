package artifacts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

func TestGenerateArtifacts(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Network.GenesisTime = 1_700_000_100
	cfg.Consensus.ValidatorCount = 1
	cfg.Consensus.Mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

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
}
