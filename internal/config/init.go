package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// DefaultInitConfig returns the deterministic starter config written by init.
func DefaultInitConfig() *Config {
	return &Config{
		Network: NetworkConfig{
			Name:        DefaultNetworkName,
			ChainID:     DefaultChainID,
			GenesisTime: 0,
		},
		Execution: ExecutionConfig{
			GasLimit:      DefaultGasLimit,
			ExtraData:     "0x",
			BaseFeePerGas: DefaultBaseFeePerGas,
			Contracts:     nil,
			Prefund:       map[string]string{},
		},
		Consensus: ConsensusConfig{
			Fork:                   DefaultFork,
			PresetBase:             DefaultPresetBase,
			ValidatorCount:         DefaultValidatorCount,
			ValidatorBalanceGwei:   DefaultValidatorBalanceGwei,
			Mnemonic:               "",
			WithdrawalAddress:      DefaultWithdrawalAddress,
			WithdrawalPrefix:       DefaultWithdrawalPrefix,
			DepositContractAddress: DefaultDepositContractAddress,
			OutputJSON:             new(true),
		},
	}
}

// ParsePrefundEntries converts repeated ADDRESS=AMOUNT CLI values into a prefund map.
func ParsePrefundEntries(entries []string) (map[string]string, error) {
	prefund := make(map[string]string, len(entries))
	for _, entry := range entries {
		address, amount, ok := strings.Cut(entry, "=")
		if !ok || address == "" || amount == "" {
			return nil, fmt.Errorf("prefund value %q must be ADDRESS=AMOUNT", entry)
		}
		if _, exists := prefund[address]; exists {
			return nil, fmt.Errorf("duplicate prefund address %s", address)
		}
		if err := ValidateAddress("execution.prefund address", address); err != nil {
			return nil, err
		}
		if _, err := ParseBigInt(amount); err != nil {
			return nil, fmt.Errorf("execution.prefund[%s]: %w", address, err)
		}
		prefund[address] = amount
	}
	return prefund, nil
}

// ValidateInit validates a config template while allowing genesis_time: 0.
func (c *Config) ValidateInit() error {
	clone := *c
	if clone.Network.GenesisTime == 0 {
		clone.Network.GenesisTime = 1
	}
	return clone.Validate()
}

// RenderInitYAML renders the starter genesis.yaml with stable field and map order.
func RenderInitYAML(c *Config) string {
	var b strings.Builder

	fmt.Fprintln(&b, "network:")
	fmt.Fprintf(&b, "  name: %s\n", yamlPlainOrQuote(c.Network.Name))
	fmt.Fprintf(&b, "  chain_id: %d\n", c.Network.ChainID)
	fmt.Fprintln(&b, "  # Leave empty or set 0 to use current time + 60 seconds.")
	fmt.Fprintf(&b, "  genesis_time: %d\n\n", c.Network.GenesisTime)

	fmt.Fprintln(&b, "execution:")
	fmt.Fprintf(&b, "  gas_limit: %d\n", c.Execution.GasLimit)
	fmt.Fprintf(&b, "  extra_data: %s\n", yamlQuote(c.Execution.ExtraData))
	fmt.Fprintf(&b, "  base_fee_per_gas: %s\n", yamlQuote(c.Execution.BaseFeePerGas))
	fmt.Fprintf(&b, "  # Optional predeploy profiles: %s.\n", SupportedExecutionContractProfiles())
	if len(c.Execution.Contracts) == 0 {
		fmt.Fprintln(&b, "  contracts: []")
	} else {
		fmt.Fprintln(&b, "  contracts:")
		for _, profile := range c.Execution.Contracts {
			fmt.Fprintf(&b, "    - %s\n", yamlPlainOrQuote(profile))
		}
	}
	if len(c.Execution.Prefund) == 0 {
		fmt.Fprintln(&b, "  prefund: {}")
	} else {
		fmt.Fprintln(&b, "  prefund:")
		addresses := make([]string, 0, len(c.Execution.Prefund))
		for address := range c.Execution.Prefund {
			addresses = append(addresses, address)
		}
		sort.Strings(addresses)
		for _, address := range addresses {
			fmt.Fprintf(&b, "    %s: %s\n", yamlQuote(address), yamlQuote(c.Execution.Prefund[address]))
		}
	}

	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "consensus:")
	fmt.Fprintf(&b, "  fork: %s\n", yamlPlainOrQuote(c.Consensus.Fork))
	fmt.Fprintf(&b, "  preset_base: %s\n", yamlPlainOrQuote(c.Consensus.PresetBase))
	fmt.Fprintf(&b, "  validator_count: %d\n", c.Consensus.ValidatorCount)
	fmt.Fprintf(&b, "  validator_balance_gwei: %d\n", c.Consensus.ValidatorBalanceGwei)
	fmt.Fprintln(&b, "  # Leave empty to generate a new 24-word mnemonic.")
	fmt.Fprintf(&b, "  mnemonic: %s\n", yamlQuote(c.Consensus.Mnemonic))
	fmt.Fprintf(&b, "  withdrawal_address: %s\n", yamlQuote(c.Consensus.WithdrawalAddress))
	fmt.Fprintf(&b, "  withdrawal_prefix: %s\n", yamlQuote(c.Consensus.WithdrawalPrefix))
	fmt.Fprintf(&b, "  deposit_contract_address: %s\n", yamlQuote(c.Consensus.DepositContractAddress))
	fmt.Fprintf(&b, "  output_json: %t\n", c.OutputJSONEnabled())
	fmt.Fprintln(&b, "  # Optional 4-byte fork version overrides.")
	fmt.Fprintln(&b, "  # fork_versions:")
	fmt.Fprintln(&b, "  #   genesis: \"0x00007e7e\"")
	fmt.Fprintln(&b, "  #   altair: \"0x01007e7e\"")
	fmt.Fprintln(&b, "  #   bellatrix: \"0x02007e7e\"")
	fmt.Fprintln(&b, "  #   capella: \"0x03007e7e\"")
	fmt.Fprintln(&b, "  #   deneb: \"0x04007e7e\"")
	fmt.Fprintln(&b, "  #   electra: \"0x05007e7e\"")
	fmt.Fprintln(&b, "  #   fulu: \"0x06007e7e\"")

	return b.String()
}

func yamlPlainOrQuote(value string) string {
	if isPlainYAMLString(value) {
		return value
	}
	return yamlQuote(value)
}

func yamlQuote(value string) string {
	return strconv.Quote(value)
}

func isPlainYAMLString(value string) bool {
	if value == "" {
		return false
	}
	for i, r := range value {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' && r != '.' {
			return false
		}
	}
	switch strings.ToLower(value) {
	case "true", "false", "null", "yes", "no", "on", "off":
		return false
	}
	return true
}
