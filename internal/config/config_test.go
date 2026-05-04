package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFileAppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.yaml")
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path, time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Network.Name != DefaultNetworkName {
		t.Fatalf("network name = %q", cfg.Network.Name)
	}
	if cfg.Network.ChainID != DefaultChainID {
		t.Fatalf("chain id = %d", cfg.Network.ChainID)
	}
	if cfg.Network.GenesisTime != 1_700_000_060 {
		t.Fatalf("genesis time = %d", cfg.Network.GenesisTime)
	}
	if !cfg.OutputJSONEnabled() {
		t.Fatal("expected output_json default to true")
	}
}

func TestLoadFileRejectsInvalidPrefundAddress(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.yaml")
	data := []byte(`execution:
  prefund:
    "0x1234": "1"
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(path, time.Unix(1_700_000_000, 0))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestForkVersionDerivation(t *testing.T) {
	got := ForkVersion(32382, "fulu", nil)
	if got != "0x06007e7e" {
		t.Fatalf("fork version = %s", got)
	}

	override := ForkVersion(32382, "phase0", map[string]string{"genesis": "0x11223344"})
	if override != "0x11223344" {
		t.Fatalf("override = %s", override)
	}
}

func TestForkVersionOverridesAreNormalized(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.yaml")
	data := []byte(`consensus:
  fork_versions:
    Fulu: "0xAABBCCDD"
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path, time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatal(err)
	}
	got := ForkVersion(cfg.Network.ChainID, "fulu", cfg.Consensus.ForkVersions)
	if got != "0xaabbccdd" {
		t.Fatalf("normalized override = %s", got)
	}
}
