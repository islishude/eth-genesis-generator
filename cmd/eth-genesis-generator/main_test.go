package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
	"gopkg.in/yaml.v3"
)

func TestRunInitWritesDefaultConfigOnly(t *testing.T) {
	outDir := t.TempDir()
	code, stdout, stderr := runCLI("init", "--out", outDir)
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "genesis.yaml") {
		t.Fatalf("stdout = %q", stdout)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "genesis.yaml" || entries[0].IsDir() {
		t.Fatalf("unexpected init outputs: %#v", entries)
	}
	for _, relPath := range []string{
		"execution",
		"consensus",
		"validators",
		"manifest.json",
	} {
		if _, err := os.Stat(filepath.Join(outDir, relPath)); !os.IsNotExist(err) {
			t.Fatalf("init created unexpected %s: %v", relPath, err)
		}
	}

	cfg := readInitConfig(t, outDir)
	if cfg.Network.Name != appconfig.DefaultNetworkName {
		t.Fatalf("network name = %q", cfg.Network.Name)
	}
	if cfg.Network.ChainID != appconfig.DefaultChainID {
		t.Fatalf("chain id = %d", cfg.Network.ChainID)
	}
	if cfg.Network.GenesisTime != 0 {
		t.Fatalf("genesis time = %d", cfg.Network.GenesisTime)
	}
	if cfg.Execution.GasLimit != appconfig.DefaultGasLimit {
		t.Fatalf("gas limit = %d", cfg.Execution.GasLimit)
	}
	if cfg.Execution.ExtraData != "0x" {
		t.Fatalf("extra data = %q", cfg.Execution.ExtraData)
	}
	if cfg.Execution.BaseFeePerGas != appconfig.DefaultBaseFeePerGas {
		t.Fatalf("base fee per gas = %q", cfg.Execution.BaseFeePerGas)
	}
	if len(cfg.Execution.Prefund) != 1 {
		t.Fatalf("prefund len = %d", len(cfg.Execution.Prefund))
	}
	if cfg.Execution.Prefund[appconfig.DefaultPrefundAddress] != appconfig.DefaultPrefundBalanceWei {
		t.Fatalf("default prefund = %#v", cfg.Execution.Prefund)
	}
	if cfg.Consensus.Fork != appconfig.DefaultFork {
		t.Fatalf("fork = %q", cfg.Consensus.Fork)
	}
	if cfg.Consensus.PresetBase != appconfig.DefaultPresetBase {
		t.Fatalf("preset base = %q", cfg.Consensus.PresetBase)
	}
	if cfg.Consensus.ValidatorCount != appconfig.DefaultValidatorCount {
		t.Fatalf("validator count = %d", cfg.Consensus.ValidatorCount)
	}
	if cfg.Consensus.ValidatorBalanceGwei != appconfig.DefaultValidatorBalanceGwei {
		t.Fatalf("validator balance = %d", cfg.Consensus.ValidatorBalanceGwei)
	}
	if cfg.Consensus.Mnemonic != "" {
		t.Fatalf("mnemonic = %q", cfg.Consensus.Mnemonic)
	}
	if cfg.Consensus.WithdrawalAddress != appconfig.DefaultWithdrawalAddress {
		t.Fatalf("withdrawal address = %q", cfg.Consensus.WithdrawalAddress)
	}
	if cfg.Consensus.WithdrawalPrefix != appconfig.DefaultWithdrawalPrefix {
		t.Fatalf("withdrawal prefix = %q", cfg.Consensus.WithdrawalPrefix)
	}
	if cfg.Consensus.DepositContractAddress != appconfig.DefaultDepositContractAddress {
		t.Fatalf("deposit contract address = %q", cfg.Consensus.DepositContractAddress)
	}
	if !cfg.OutputJSONEnabled() {
		t.Fatal("expected output_json default to true")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestRunInitFlagOverrides(t *testing.T) {
	outDir := t.TempDir()
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	code, _, stderr := runCLI(
		"init",
		"--out", outDir,
		"--network-name", "customnet",
		"--chain-id", "4242",
		"--genesis-time", "1700000100",
		"--gas-limit", "123456",
		"--base-fee-per-gas", "42",
		"--prefund", "0x1000000000000000000000000000000000000002=10",
		"--prefund", "0x1000000000000000000000000000000000000003=20",
		"--fork", "deneb",
		"--preset-base", "minimal",
		"--validator-count", "4",
		"--validator-balance-gwei", "12345",
		"--mnemonic", mnemonic,
		"--withdrawal-address", "0x1000000000000000000000000000000000000004",
		"--withdrawal-prefix", "0x01",
		"--deposit-contract-address", "0x1000000000000000000000000000000000000005",
		"--output-json=false",
	)
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr)
	}

	cfg := readInitConfig(t, outDir)
	if cfg.Network.Name != "customnet" ||
		cfg.Network.ChainID != 4242 ||
		cfg.Network.GenesisTime != 1_700_000_100 ||
		cfg.Execution.GasLimit != 123456 ||
		cfg.Execution.BaseFeePerGas != "42" ||
		cfg.Consensus.Fork != "deneb" ||
		cfg.Consensus.PresetBase != "minimal" ||
		cfg.Consensus.ValidatorCount != 4 ||
		cfg.Consensus.ValidatorBalanceGwei != 12345 ||
		cfg.Consensus.Mnemonic != mnemonic ||
		cfg.Consensus.WithdrawalAddress != "0x1000000000000000000000000000000000000004" ||
		cfg.Consensus.WithdrawalPrefix != "0x01" ||
		cfg.Consensus.DepositContractAddress != "0x1000000000000000000000000000000000000005" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if len(cfg.Execution.Prefund) != 2 {
		t.Fatalf("prefund len = %d", len(cfg.Execution.Prefund))
	}
	if cfg.Execution.Prefund["0x1000000000000000000000000000000000000002"] != "10" {
		t.Fatalf("prefund = %#v", cfg.Execution.Prefund)
	}
	if cfg.Execution.Prefund["0x1000000000000000000000000000000000000003"] != "20" {
		t.Fatalf("prefund = %#v", cfg.Execution.Prefund)
	}
	if cfg.Execution.Prefund[appconfig.DefaultPrefundAddress] != "" {
		t.Fatalf("default prefund should be replaced: %#v", cfg.Execution.Prefund)
	}
	if cfg.OutputJSONEnabled() {
		t.Fatal("expected output_json=false")
	}
}

func TestRunInitOverwriteProtection(t *testing.T) {
	outDir := t.TempDir()
	configPath := filepath.Join(outDir, "genesis.yaml")
	if err := os.WriteFile(configPath, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runCLI("init", "--out", outDir)
	if code == 0 {
		t.Fatal("expected init to fail without --force")
	}
	if !strings.Contains(stderr, "already exists") {
		t.Fatalf("stderr = %q", stderr)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original\n" {
		t.Fatalf("file was overwritten: %q", string(data))
	}

	code, _, stderr = runCLI("init", "--out", outDir, "--network-name", "forced", "--force")
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr)
	}
	cfg := readInitConfig(t, outDir)
	if cfg.Network.Name != "forced" {
		t.Fatalf("network name = %q", cfg.Network.Name)
	}
}

func TestRunInitValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{
			name:       "bad fork",
			args:       []string{"--fork", "notafork"},
			wantStderr: "unsupported",
		},
		{
			name:       "malformed prefund",
			args:       []string{"--prefund", "0x1000000000000000000000000000000000000001"},
			wantStderr: "ADDRESS=AMOUNT",
		},
		{
			name:       "invalid prefund address",
			args:       []string{"--prefund", "0x1234=1"},
			wantStderr: "not a valid Ethereum address",
		},
		{
			name: "duplicate prefund address",
			args: []string{
				"--prefund", "0x1000000000000000000000000000000000000002=1",
				"--prefund", "0x1000000000000000000000000000000000000002=2",
			},
			wantStderr: "duplicate prefund address",
		},
		{
			name:       "invalid withdrawal address",
			args:       []string{"--withdrawal-address", "0x1234"},
			wantStderr: "consensus.withdrawal_address",
		},
		{
			name:       "invalid deposit contract address",
			args:       []string{"--deposit-contract-address", "0x1234"},
			wantStderr: "consensus.deposit_contract_address",
		},
		{
			name:       "invalid withdrawal prefix",
			args:       []string{"--withdrawal-prefix", "0x1234"},
			wantStderr: "consensus.withdrawal_prefix",
		},
		{
			name:       "zero chain ID",
			args:       []string{"--chain-id", "0"},
			wantStderr: "network.chain_id",
		},
		{
			name:       "zero gas limit",
			args:       []string{"--gas-limit", "0"},
			wantStderr: "execution.gas_limit",
		},
		{
			name:       "zero validator count",
			args:       []string{"--validator-count", "0"},
			wantStderr: "consensus.validator_count",
		},
		{
			name:       "zero validator balance",
			args:       []string{"--validator-balance-gwei", "0"},
			wantStderr: "consensus.validator_balance_gwei",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := t.TempDir()
			args := append([]string{"init", "--out", outDir}, tt.args...)
			code, _, stderr := runCLI(args...)
			if code == 0 {
				t.Fatal("expected non-zero exit code")
			}
			if !strings.Contains(stderr, tt.wantStderr) {
				t.Fatalf("stderr = %q, want %q", stderr, tt.wantStderr)
			}
			if _, err := os.Stat(filepath.Join(outDir, "genesis.yaml")); !os.IsNotExist(err) {
				t.Fatalf("genesis.yaml should not be written: %v", err)
			}
		})
	}
}

func runCLI(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func readInitConfig(t *testing.T, outDir string) appconfig.Config {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(outDir, "genesis.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg appconfig.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}
