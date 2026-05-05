package config

import (
	"os"
	"path/filepath"
	"strings"
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
	if len(cfg.Execution.Contracts) != 0 {
		t.Fatalf("execution contracts = %#v", cfg.Execution.Contracts)
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

func TestLoadFileNormalizesExecutionContractProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.yaml")
	data := []byte(`execution:
  contracts:
    - System
    - Utils
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path, time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Execution.Contracts) != 2 ||
		cfg.Execution.Contracts[0] != ExecutionContractsProfileSystem ||
		cfg.Execution.Contracts[1] != ExecutionContractsProfileUtils {
		t.Fatalf("execution contracts = %#v", cfg.Execution.Contracts)
	}
}

func TestLoadFileTreatsExecutionContractNoneAsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.yaml")
	data := []byte(`execution:
  contracts:
    - none
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path, time.Unix(1_700_000_000, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Execution.Contracts) != 0 {
		t.Fatalf("execution contracts = %#v", cfg.Execution.Contracts)
	}
}

func TestLoadFileRejectsInvalidExecutionContractProfiles(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "unsupported",
			data: `execution:
  contracts:
    - custom
`,
			want: "unsupported profile",
		},
		{
			name: "duplicate",
			data: `execution:
  contracts:
    - system
    - System
`,
			want: "duplicate profile",
		},
		{
			name: "none combined",
			data: `execution:
  contracts:
    - none
    - system
`,
			want: "cannot be combined",
		},
		{
			name: "all combined",
			data: `execution:
  contracts:
    - all
    - utils
`,
			want: "cannot be combined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "genesis.yaml")
			if err := os.WriteFile(path, []byte(tt.data), 0o644); err != nil {
				t.Fatal(err)
			}

			_, err := LoadFile(path, time.Unix(1_700_000_000, 0))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want %q", err, tt.want)
			}
		})
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
