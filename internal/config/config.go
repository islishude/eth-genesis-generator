package config

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	bip39 "github.com/tyler-smith/go-bip39"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultNetworkName is the network name used when genesis.yaml omits one.
	DefaultNetworkName = "devnet"

	// DefaultChainID is the execution and deposit chain ID for local devnets.
	DefaultChainID = uint64(32382)

	// DefaultGasLimit is the execution genesis gas limit.
	DefaultGasLimit = uint64(60_000_000)

	// DefaultBaseFeePerGas is the London base fee encoded as a decimal string.
	DefaultBaseFeePerGas = "1000000000"

	// DefaultFork is the consensus fork activated at genesis.
	DefaultFork = "fulu"

	// DefaultPresetBase selects the eth-beacon-genesis preset family.
	DefaultPresetBase = "mainnet"

	// DefaultValidatorCount is the number of pre-filled validators.
	DefaultValidatorCount = uint64(64)

	// DefaultValidatorBalanceGwei is the effective validator balance in gwei.
	DefaultValidatorBalanceGwei = uint64(32_000_000_000)

	// DefaultWithdrawalAddress receives validator withdrawal credentials.
	DefaultWithdrawalAddress = "0x1000000000000000000000000000000000000001"

	// DefaultWithdrawalPrefix selects compounding withdrawal credentials.
	DefaultWithdrawalPrefix = "0x02"

	// DefaultDepositContractAddress is referenced by EL and CL configs but is not predeployed.
	DefaultDepositContractAddress = "0x4242424242424242424242424242424242424242"

	// ExecutionContractsProfileSystem predeploys Ethereum system contracts.
	ExecutionContractsProfileSystem = "system"

	// ExecutionContractsProfileUtils predeploys auxiliary local devnet contracts.
	ExecutionContractsProfileUtils = "utils"

	// ExecutionContractsProfileAll predeploys all bundled execution contracts.
	ExecutionContractsProfileAll = "all"

	// ExecutionContractsProfileNone keeps execution genesis free of bundled contracts.
	ExecutionContractsProfileNone = "none"
)

// Config is the root user-facing generator configuration.
type Config struct {
	// Network contains chain identity and genesis timing shared across layers.
	Network NetworkConfig `yaml:"network"`
	// Execution contains execution-layer genesis settings.
	Execution ExecutionConfig `yaml:"execution"`
	// Consensus contains consensus-layer genesis and validator settings.
	Consensus ConsensusConfig `yaml:"consensus"`
}

// NetworkConfig defines network identity and the genesis timestamp.
type NetworkConfig struct {
	// Name is written into consensus CONFIG_NAME and manifest metadata.
	Name string `yaml:"name"`
	// ChainID is used for execution ChainID and deposit chain/network IDs.
	ChainID uint64 `yaml:"chain_id"`
	// GenesisTime is a Unix timestamp. A zero value defaults to now + 60 seconds.
	GenesisTime uint64 `yaml:"genesis_time,omitempty"`
}

// ExecutionConfig defines fields used to build execution/genesis.json.
type ExecutionConfig struct {
	// GasLimit is the genesis block gas limit.
	GasLimit uint64 `yaml:"gas_limit"`
	// ExtraData is 0x-prefixed hex capped at the execution-layer 32-byte limit.
	ExtraData string `yaml:"extra_data"`
	// BaseFeePerGas is a non-negative decimal integer string.
	BaseFeePerGas string `yaml:"base_fee_per_gas"`
	// Contracts selects optional bundled execution contract predeploy profiles.
	Contracts []string `yaml:"contracts,omitempty"`
	// Prefund maps 0x-prefixed addresses to wei balances encoded as decimal strings.
	Prefund map[string]string `yaml:"prefund,omitempty"`
}

// ConsensusConfig defines fields used for config.yaml, mnemonics.yaml, and beacon genesis.
type ConsensusConfig struct {
	// Fork is the highest consensus fork activated at genesis.
	Fork string `yaml:"fork"`
	// PresetBase selects the consensus preset consumed by eth-beacon-genesis.
	PresetBase string `yaml:"preset_base"`
	// ValidatorCount is the number of pre-filled validators to generate.
	ValidatorCount uint64 `yaml:"validator_count"`
	// ValidatorBalanceGwei is the initial balance assigned to each validator.
	ValidatorBalanceGwei uint64 `yaml:"validator_balance_gwei"`
	// Mnemonic is optional; an empty value generates a new 24-word mnemonic.
	Mnemonic string `yaml:"mnemonic,omitempty"`
	// WithdrawalAddress is encoded into generated validator withdrawal credentials.
	WithdrawalAddress string `yaml:"withdrawal_address"`
	// WithdrawalPrefix is a one-byte 0x-prefixed withdrawal credential prefix.
	WithdrawalPrefix string `yaml:"withdrawal_prefix"`
	// DepositContractAddress is referenced by both layers; v1 does not predeploy bytecode.
	DepositContractAddress string `yaml:"deposit_contract_address"`
	// OutputJSON controls whether consensus/genesis.json is emitted beside genesis.ssz.
	OutputJSON *bool `yaml:"output_json,omitempty"`
	// ForkVersions optionally overrides derived 4-byte fork versions by fork name.
	ForkVersions map[string]string `yaml:"fork_versions,omitempty"`
}

// LoadFile reads a YAML config file, applies defaults, and validates the result.
func LoadFile(path string, now time.Time) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if len(strings.TrimSpace(string(data))) > 0 {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	}

	cfg.ApplyDefaults(now)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ApplyDefaults fills zero-valued fields with deterministic project defaults.
func (c *Config) ApplyDefaults(now time.Time) {
	if c.Network.Name == "" {
		c.Network.Name = DefaultNetworkName
	}
	if c.Network.ChainID == 0 {
		c.Network.ChainID = DefaultChainID
	}
	if c.Network.GenesisTime == 0 {
		c.Network.GenesisTime = uint64(now.Unix()) + 60
	}
	if c.Execution.GasLimit == 0 {
		c.Execution.GasLimit = DefaultGasLimit
	}
	if c.Execution.ExtraData == "" {
		c.Execution.ExtraData = "0x"
	}
	if c.Execution.BaseFeePerGas == "" {
		c.Execution.BaseFeePerGas = DefaultBaseFeePerGas
	}
	if c.Execution.Prefund == nil {
		c.Execution.Prefund = map[string]string{}
	}
	if c.Consensus.Fork == "" {
		c.Consensus.Fork = DefaultFork
	}
	c.Consensus.Fork = strings.ToLower(c.Consensus.Fork)
	if c.Consensus.PresetBase == "" {
		c.Consensus.PresetBase = DefaultPresetBase
	}
	if c.Consensus.ValidatorCount == 0 {
		c.Consensus.ValidatorCount = DefaultValidatorCount
	}
	if c.Consensus.ValidatorBalanceGwei == 0 {
		c.Consensus.ValidatorBalanceGwei = DefaultValidatorBalanceGwei
	}
	if c.Consensus.WithdrawalAddress == "" {
		c.Consensus.WithdrawalAddress = DefaultWithdrawalAddress
	}
	if c.Consensus.WithdrawalPrefix == "" {
		c.Consensus.WithdrawalPrefix = DefaultWithdrawalPrefix
	}
	if c.Consensus.DepositContractAddress == "" {
		c.Consensus.DepositContractAddress = DefaultDepositContractAddress
	}
	if c.Consensus.OutputJSON == nil {
		c.Consensus.OutputJSON = new(true)
	}
	if c.Consensus.ForkVersions != nil {
		normalized := make(map[string]string, len(c.Consensus.ForkVersions))
		for key, value := range c.Consensus.ForkVersions {
			normalized[strings.ToLower(key)] = strings.ToLower(value)
		}
		c.Consensus.ForkVersions = normalized
	}
}

// Validate checks semantic constraints that cannot be expressed by YAML tags.
func (c *Config) Validate() error {
	if c.Network.ChainID == 0 {
		return fmt.Errorf("network.chain_id must be greater than zero")
	}
	if c.Network.GenesisTime == 0 {
		return fmt.Errorf("network.genesis_time must be greater than zero")
	}
	if c.Execution.GasLimit == 0 {
		return fmt.Errorf("execution.gas_limit must be greater than zero")
	}
	if _, err := DecodeHexBytes(c.Execution.ExtraData); err != nil {
		return fmt.Errorf("execution.extra_data: %w", err)
	}
	extra, _ := DecodeHexBytes(c.Execution.ExtraData)
	if len(extra) > 32 {
		return fmt.Errorf("execution.extra_data is %d bytes, max is 32", len(extra))
	}
	if _, err := ParseBigInt(c.Execution.BaseFeePerGas); err != nil {
		return fmt.Errorf("execution.base_fee_per_gas: %w", err)
	}
	contracts, err := NormalizeExecutionContractProfiles(c.Execution.Contracts)
	if err != nil {
		return err
	}
	c.Execution.Contracts = contracts

	addresses := make([]string, 0, len(c.Execution.Prefund))
	for address := range c.Execution.Prefund {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	for _, address := range addresses {
		if err := ValidateAddress("execution.prefund address", address); err != nil {
			return err
		}
		if _, err := ParseBigInt(c.Execution.Prefund[address]); err != nil {
			return fmt.Errorf("execution.prefund[%s]: %w", address, err)
		}
	}

	if _, ok := ForkIndex(c.Consensus.Fork); !ok {
		return fmt.Errorf("consensus.fork %q is unsupported; supported forks: %s", c.Consensus.Fork, strings.Join(SupportedForks(), ", "))
	}
	if c.Consensus.ValidatorCount == 0 {
		return fmt.Errorf("consensus.validator_count must be greater than zero")
	}
	if c.Consensus.ValidatorBalanceGwei == 0 {
		return fmt.Errorf("consensus.validator_balance_gwei must be greater than zero")
	}
	if c.Consensus.Mnemonic != "" && !bip39.IsMnemonicValid(c.Consensus.Mnemonic) {
		return fmt.Errorf("consensus.mnemonic is not a valid BIP-39 mnemonic")
	}
	if err := ValidateAddress("consensus.withdrawal_address", c.Consensus.WithdrawalAddress); err != nil {
		return err
	}
	if err := ValidateAddress("consensus.deposit_contract_address", c.Consensus.DepositContractAddress); err != nil {
		return err
	}
	if err := ValidateHexFixed("consensus.withdrawal_prefix", c.Consensus.WithdrawalPrefix, 1); err != nil {
		return err
	}
	for name, version := range c.Consensus.ForkVersions {
		if _, ok := forkFieldName(name); !ok {
			return fmt.Errorf("consensus.fork_versions has unsupported key %q", name)
		}
		if err := ValidateHexFixed("consensus.fork_versions."+name, version, 4); err != nil {
			return err
		}
	}

	return nil
}

// OutputJSONEnabled returns true unless consensus.output_json is explicitly false.
func (c *Config) OutputJSONEnabled() bool {
	return c.Consensus.OutputJSON == nil || *c.Consensus.OutputJSON
}

// ValidateAddress checks that a value is a 0x-prefixed Ethereum address.
func ValidateAddress(name string, address string) error {
	if !strings.HasPrefix(address, "0x") {
		return fmt.Errorf("%s must be a 0x-prefixed Ethereum address", name)
	}
	if !common.IsHexAddress(address) {
		return fmt.Errorf("%s %q is not a valid Ethereum address", name, address)
	}
	return nil
}

// ValidateHexFixed checks that a value is 0x-prefixed hex with exactly wantBytes bytes.
func ValidateHexFixed(name string, value string, wantBytes int) error {
	decoded, err := DecodeHexBytes(value)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if len(decoded) != wantBytes {
		return fmt.Errorf("%s must be %d bytes, got %d", name, wantBytes, len(decoded))
	}
	return nil
}

// DecodeHexBytes decodes a 0x-prefixed hex string and rejects odd-length input.
func DecodeHexBytes(value string) ([]byte, error) {
	if !strings.HasPrefix(value, "0x") {
		return nil, fmt.Errorf("must be 0x-prefixed hex")
	}
	if len(value)%2 != 0 {
		return nil, fmt.Errorf("hex string must have an even number of digits")
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "0x"))
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

// ParseBigInt parses a non-negative base-10 integer string.
func ParseBigInt(value string) (*big.Int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("must not be empty")
	}
	n, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, fmt.Errorf("must be a base-10 integer string")
	}
	if n.Sign() < 0 {
		return nil, fmt.Errorf("must not be negative")
	}
	return n, nil
}
