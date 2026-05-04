package consensus

import (
	"fmt"
	"strings"

	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

// RenderConfig renders the consensus config.yaml consumed by eth-beacon-genesis.
func RenderConfig(cfg *appconfig.Config) (string, error) {
	activeForks, err := appconfig.ActiveForks(cfg.Consensus.Fork)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "PRESET_BASE: %q\n", cfg.Consensus.PresetBase)
	fmt.Fprintf(&b, "CONFIG_NAME: %q\n\n", cfg.Network.Name)

	fmt.Fprintln(&b, "# Genesis")
	fmt.Fprintf(&b, "MIN_GENESIS_ACTIVE_VALIDATOR_COUNT: %d\n", cfg.Consensus.ValidatorCount)
	fmt.Fprintf(&b, "MIN_GENESIS_TIME: %d\n", cfg.Network.GenesisTime)
	fmt.Fprintln(&b, "GENESIS_DELAY: 0")
	fmt.Fprintf(&b, "DEPOSIT_CHAIN_ID: %d\n", cfg.Network.ChainID)
	fmt.Fprintf(&b, "DEPOSIT_NETWORK_ID: %d\n", cfg.Network.ChainID)
	fmt.Fprintf(&b, "DEPOSIT_CONTRACT_ADDRESS: %q\n\n", cfg.Consensus.DepositContractAddress)

	fmt.Fprintln(&b, "# Forking")
	for _, fork := range activeForks {
		// Every fork up to the configured target is activated at epoch 0 for devnets.
		fmt.Fprintf(
			&b,
			"%s: %s\n",
			appconfig.ForkVersionField(fork),
			appconfig.ForkVersion(cfg.Network.ChainID, fork, cfg.Consensus.ForkVersions),
		)
		if epochField := appconfig.ForkEpochField(fork); epochField != "" {
			fmt.Fprintf(&b, "%s: 0\n", epochField)
		}
	}

	return b.String(), nil
}

// RenderMnemonics renders validator input for eth-beacon-genesis.
func RenderMnemonics(cfg *appconfig.Config, mnemonic string) string {
	return fmt.Sprintf(`- mnemonic: %q
  start: 0
  count: %d
  balance: %d
  wd_address: %q
  wd_prefix: %q
  status: 0
`, mnemonic, cfg.Consensus.ValidatorCount, cfg.Consensus.ValidatorBalanceGwei, cfg.Consensus.WithdrawalAddress, cfg.Consensus.WithdrawalPrefix)
}
