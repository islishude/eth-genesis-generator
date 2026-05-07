package consensus

import (
	"strings"
	"testing"
	"time"

	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
)

func TestRenderConfigActivatesFuluAtGenesis(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Network.GenesisTime = 1_700_000_100

	out, err := RenderConfig(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		`PRESET_BASE: "mainnet"`,
		"TERMINAL_TOTAL_DIFFICULTY: 0",
		"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT: 64",
		"GENESIS_DELAY: 0",
		"FULU_FORK_VERSION: 0x06007e7e",
		"FULU_FORK_EPOCH: 0",
		"GLOAS_FORK_EPOCH: 18446744073709551615",
		"SECONDS_PER_SLOT: 12",
		"SLOT_DURATION_MS: 12000",
		"SECONDS_PER_ETH1_BLOCK: 14",
		"MIN_VALIDATOR_WITHDRAWABILITY_DELAY: 256",
		"SHARD_COMMITTEE_PERIOD: 256",
		"ETH1_FOLLOW_DISTANCE: 2048",
		"INACTIVITY_SCORE_BIAS: 4",
		"PROPOSER_SCORE_BOOST: 40",
		"MAX_PAYLOAD_SIZE: 10485760",
		"MAX_BLOBS_PER_BLOCK_ELECTRA: 9",
		"BLOB_SCHEDULE: []",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered config missing %q:\n%s", want, out)
		}
	}
}

func TestRenderConfigDisablesForksAfterTarget(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Consensus.Fork = "deneb"

	out, err := RenderConfig(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"DENEB_FORK_EPOCH: 0",
		"ELECTRA_FORK_EPOCH: 18446744073709551615",
		"FULU_FORK_EPOCH: 18446744073709551615",
		"GLOAS_FORK_EPOCH: 18446744073709551615",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered config missing %q:\n%s", want, out)
		}
	}
}

func TestRenderConfigMinimalTiming(t *testing.T) {
	var cfg appconfig.Config
	cfg.ApplyDefaults(time.Unix(1_700_000_000, 0))
	cfg.Consensus.PresetBase = "minimal"

	out, err := RenderConfig(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		`PRESET_BASE: "minimal"`,
		"SECONDS_PER_SLOT: 6",
		"SLOT_DURATION_MS: 6000",
		"SHARD_COMMITTEE_PERIOD: 64",
		"ETH1_FOLLOW_DISTANCE: 16",
		"MIN_EPOCHS_FOR_BLOCK_REQUESTS: 272",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered config missing %q:\n%s", want, out)
		}
	}
}
