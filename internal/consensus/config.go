package consensus

import (
	"fmt"
	"strings"

	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

const farFutureEpoch = uint64(^uint64(0))

type chainSpecValues struct {
	SecondsPerSlot                      uint64
	SlotDurationMS                      uint64
	ShardCommitteePeriod                uint64
	Eth1FollowDistance                  uint64
	MinEpochsForBlockRequests           uint64
	MinBuilderWithdrawabilityDelay      uint64
	MinPerEpochChurnLimit               uint64
	ChurnLimitQuotient                  uint64
	MaxPerEpochActivationChurnLimit     uint64
	MinPerEpochChurnLimitElectra        uint64
	MaxPerEpochActivationExitChurnLimit uint64
}

// RenderConfig renders the consensus config.yaml consumed by eth-beacon-genesis and clients.
func RenderConfig(cfg *appconfig.Config) (string, error) {
	activeForkIndex, ok := appconfig.ForkIndex(cfg.Consensus.Fork)
	if !ok {
		return "", fmt.Errorf("unsupported fork %q", cfg.Consensus.Fork)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "PRESET_BASE: %q\n", cfg.Consensus.PresetBase)
	fmt.Fprintf(&b, "CONFIG_NAME: %q\n\n", cfg.Network.Name)

	fmt.Fprintln(&b, "# Transition")
	fmt.Fprintln(&b, "TERMINAL_TOTAL_DIFFICULTY: 0")
	fmt.Fprintln(&b, "TERMINAL_BLOCK_HASH: 0x0000000000000000000000000000000000000000000000000000000000000000")
	fmt.Fprintf(&b, "TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH: %d\n\n", farFutureEpoch)

	fmt.Fprintln(&b, "# Genesis")
	fmt.Fprintf(&b, "MIN_GENESIS_ACTIVE_VALIDATOR_COUNT: %d\n", cfg.Consensus.ValidatorCount)
	fmt.Fprintf(&b, "MIN_GENESIS_TIME: %d\n", cfg.Network.GenesisTime)
	fmt.Fprintf(
		&b,
		"GENESIS_FORK_VERSION: %s\n",
		appconfig.ForkVersion(cfg.Network.ChainID, "phase0", cfg.Consensus.ForkVersions),
	)
	fmt.Fprintln(&b, "GENESIS_DELAY: 0")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "# Forking")
	for forkIndex, fork := range appconfig.SupportedForks() {
		if fork == "phase0" {
			continue
		}
		epoch := farFutureEpoch
		if forkIndex <= activeForkIndex {
			epoch = 0
		}
		fmt.Fprintf(&b, "%s: %s\n", appconfig.ForkVersionField(fork), appconfig.ForkVersion(cfg.Network.ChainID, fork, cfg.Consensus.ForkVersions))
		fmt.Fprintf(&b, "%s: %d\n", appconfig.ForkEpochField(fork), epoch)
	}
	fmt.Fprintf(&b, "GLOAS_FORK_VERSION: %s\n", derivedForkVersion(cfg.Network.ChainID, 0x07))
	fmt.Fprintf(&b, "GLOAS_FORK_EPOCH: %d\n", farFutureEpoch)

	writeChainSpec(&b, cfg)

	return b.String(), nil
}

func writeChainSpec(b *strings.Builder, cfg *appconfig.Config) {
	values := chainSpecForPreset(cfg.Consensus.PresetBase)

	fmt.Fprintln(b, "# Time parameters")
	fmt.Fprintf(b, "SECONDS_PER_SLOT: %d\n", values.SecondsPerSlot)
	fmt.Fprintf(b, "SLOT_DURATION_MS: %d\n", values.SlotDurationMS)
	fmt.Fprintln(b, "SECONDS_PER_ETH1_BLOCK: 14")
	fmt.Fprintln(b, "MIN_VALIDATOR_WITHDRAWABILITY_DELAY: 256")
	fmt.Fprintf(b, "SHARD_COMMITTEE_PERIOD: %d\n", values.ShardCommitteePeriod)
	fmt.Fprintf(b, "ETH1_FOLLOW_DISTANCE: %d\n", values.Eth1FollowDistance)
	fmt.Fprintln(b, "PROPOSER_REORG_CUTOFF_BPS: 1667")
	fmt.Fprintln(b, "ATTESTATION_DUE_BPS: 3333")
	fmt.Fprintln(b, "AGGREGATE_DUE_BPS: 6667")
	fmt.Fprintln(b, "SYNC_MESSAGE_DUE_BPS: 3333")
	fmt.Fprintln(b, "CONTRIBUTION_DUE_BPS: 6667")
	fmt.Fprintf(b, "MIN_BUILDER_WITHDRAWABILITY_DELAY: %d\n", values.MinBuilderWithdrawabilityDelay)
	fmt.Fprintln(b, "ATTESTATION_DUE_BPS_GLOAS: 2500")
	fmt.Fprintln(b, "AGGREGATE_DUE_BPS_GLOAS: 5000")
	fmt.Fprintln(b, "SYNC_MESSAGE_DUE_BPS_GLOAS: 2500")
	fmt.Fprintln(b, "CONTRIBUTION_DUE_BPS_GLOAS: 5000")
	fmt.Fprintln(b, "PAYLOAD_ATTESTATION_DUE_BPS: 7500")
	fmt.Fprintln(b, "VIEW_FREEZE_CUTOFF_BPS: 7500")
	fmt.Fprintln(b, "INCLUSION_LIST_SUBMISSION_DUE_BPS: 6667")
	fmt.Fprintln(b, "PROPOSER_INCLUSION_LIST_CUTOFF_BPS: 9167")
	fmt.Fprintln(b)

	fmt.Fprintln(b, "# Validator cycle")
	fmt.Fprintln(b, "INACTIVITY_SCORE_BIAS: 4")
	fmt.Fprintln(b, "INACTIVITY_SCORE_RECOVERY_RATE: 16")
	fmt.Fprintln(b, "EJECTION_BALANCE: 16000000000")
	fmt.Fprintf(b, "MIN_PER_EPOCH_CHURN_LIMIT: %d\n", values.MinPerEpochChurnLimit)
	fmt.Fprintf(b, "CHURN_LIMIT_QUOTIENT: %d\n", values.ChurnLimitQuotient)
	fmt.Fprintf(b, "MAX_PER_EPOCH_ACTIVATION_CHURN_LIMIT: %d\n", values.MaxPerEpochActivationChurnLimit)
	fmt.Fprintf(b, "MIN_PER_EPOCH_CHURN_LIMIT_ELECTRA: %d\n", values.MinPerEpochChurnLimitElectra)
	fmt.Fprintf(b, "MAX_PER_EPOCH_ACTIVATION_EXIT_CHURN_LIMIT: %d\n\n", values.MaxPerEpochActivationExitChurnLimit)

	fmt.Fprintln(b, "# Fork choice")
	fmt.Fprintln(b, "PROPOSER_SCORE_BOOST: 40")
	fmt.Fprintln(b, "REORG_HEAD_WEIGHT_THRESHOLD: 20")
	fmt.Fprintln(b, "REORG_PARENT_WEIGHT_THRESHOLD: 160")
	fmt.Fprintln(b, "REORG_MAX_EPOCHS_SINCE_FINALIZATION: 2")
	fmt.Fprintln(b)

	fmt.Fprintln(b, "# Deposit contract")
	fmt.Fprintf(b, "DEPOSIT_CHAIN_ID: %d\n", cfg.Network.ChainID)
	fmt.Fprintf(b, "DEPOSIT_NETWORK_ID: %d\n", cfg.Network.ChainID)
	fmt.Fprintf(b, "DEPOSIT_CONTRACT_ADDRESS: %q\n\n", cfg.Consensus.DepositContractAddress)

	fmt.Fprintln(b, "# Networking")
	fmt.Fprintln(b, "MAX_PAYLOAD_SIZE: 10485760")
	fmt.Fprintln(b, "MAX_REQUEST_BLOCKS: 1024")
	fmt.Fprintln(b, "EPOCHS_PER_SUBNET_SUBSCRIPTION: 256")
	fmt.Fprintf(b, "MIN_EPOCHS_FOR_BLOCK_REQUESTS: %d\n", values.MinEpochsForBlockRequests)
	fmt.Fprintln(b, "ATTESTATION_PROPAGATION_SLOT_RANGE: 32")
	fmt.Fprintln(b, "MAXIMUM_GOSSIP_CLOCK_DISPARITY: 500")
	fmt.Fprintln(b, "MESSAGE_DOMAIN_INVALID_SNAPPY: 0x00000000")
	fmt.Fprintln(b, "MESSAGE_DOMAIN_VALID_SNAPPY: 0x01000000")
	fmt.Fprintln(b, "SUBNETS_PER_NODE: 2")
	fmt.Fprintln(b, "ATTESTATION_SUBNET_COUNT: 64")
	fmt.Fprintln(b, "ATTESTATION_SUBNET_EXTRA_BITS: 0")
	fmt.Fprintln(b, "ATTESTATION_SUBNET_PREFIX_BITS: 6")
	fmt.Fprintln(b, "MAX_REQUEST_BLOCKS_DENEB: 128")
	fmt.Fprintln(b, "MIN_EPOCHS_FOR_BLOB_SIDECARS_REQUESTS: 4096")
	fmt.Fprintln(b, "BLOB_SIDECAR_SUBNET_COUNT: 6")
	fmt.Fprintln(b, "MAX_BLOBS_PER_BLOCK: 6")
	fmt.Fprintln(b, "MAX_REQUEST_BLOB_SIDECARS: 768")
	fmt.Fprintln(b, "BLOB_SIDECAR_SUBNET_COUNT_ELECTRA: 9")
	fmt.Fprintln(b, "MAX_BLOBS_PER_BLOCK_ELECTRA: 9")
	fmt.Fprintln(b, "MAX_REQUEST_BLOB_SIDECARS_ELECTRA: 1152")
	fmt.Fprintln(b, "NUMBER_OF_CUSTODY_GROUPS: 128")
	fmt.Fprintln(b, "DATA_COLUMN_SIDECAR_SUBNET_COUNT: 128")
	fmt.Fprintln(b, "MAX_REQUEST_DATA_COLUMN_SIDECARS: 16384")
	fmt.Fprintln(b, "SAMPLES_PER_SLOT: 8")
	fmt.Fprintln(b, "CUSTODY_REQUIREMENT: 4")
	fmt.Fprintln(b, "VALIDATOR_CUSTODY_REQUIREMENT: 8")
	fmt.Fprintln(b, "BALANCE_PER_ADDITIONAL_CUSTODY_GROUP: 32000000000")
	fmt.Fprintln(b, "MIN_EPOCHS_FOR_DATA_COLUMN_SIDECARS_REQUESTS: 4096")
	fmt.Fprintln(b, "MAX_REQUEST_PAYLOADS: 128")
	fmt.Fprintln(b, "EPOCHS_PER_SHUFFLING_PHASE: 256")
	fmt.Fprintln(b, "PROPOSER_SELECTION_GAP: 2")
	fmt.Fprintln(b, "MAX_REQUEST_INCLUSION_LIST: 16")
	fmt.Fprintln(b, "MAX_BYTES_PER_INCLUSION_LIST: 8192")
	fmt.Fprintln(b)

	fmt.Fprintln(b, "# Blob scheduling")
	fmt.Fprintln(b, "BLOB_SCHEDULE: []")
}

func chainSpecForPreset(presetBase string) chainSpecValues {
	//  https://github.com/ethereum/consensus-specs/blob/master/configs/minimal.yaml
	if strings.EqualFold(presetBase, "minimal") {
		return chainSpecValues{
			SecondsPerSlot:                      6,
			SlotDurationMS:                      6000,
			ShardCommitteePeriod:                64,
			Eth1FollowDistance:                  16,
			MinEpochsForBlockRequests:           272,
			MinBuilderWithdrawabilityDelay:      2,
			MinPerEpochChurnLimit:               2,
			ChurnLimitQuotient:                  32,
			MaxPerEpochActivationChurnLimit:     4,
			MinPerEpochChurnLimitElectra:        64000000000,
			MaxPerEpochActivationExitChurnLimit: 128000000000,
		}
	}

	//  https://github.com/ethereum/consensus-specs/blob/master/configs/mainnet.yaml
	return chainSpecValues{
		SecondsPerSlot:                      12,
		SlotDurationMS:                      12000,
		ShardCommitteePeriod:                256,
		Eth1FollowDistance:                  2048,
		MinEpochsForBlockRequests:           33024,
		MinBuilderWithdrawabilityDelay:      4096,
		MinPerEpochChurnLimit:               4,
		ChurnLimitQuotient:                  65536,
		MaxPerEpochActivationChurnLimit:     8,
		MinPerEpochChurnLimitElectra:        128000000000,
		MaxPerEpochActivationExitChurnLimit: 256000000000,
	}
}

func derivedForkVersion(chainID uint64, forkIndex uint32) string {
	value := (forkIndex << 24) | uint32(chainID&0x00ff_ffff)
	return fmt.Sprintf("0x%08x", value)
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
