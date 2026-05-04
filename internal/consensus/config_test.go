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
		"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT: 64",
		"GENESIS_DELAY: 0",
		"FULU_FORK_VERSION: 0x06007e7e",
		"FULU_FORK_EPOCH: 0",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered config missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "GLOAS_FORK_EPOCH") {
		t.Fatalf("gloas should not be active:\n%s", out)
	}
}
