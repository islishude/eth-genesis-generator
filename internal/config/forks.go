package config

import (
	"fmt"
	"strings"
)

const (
	forkPhase0       = "phase0"
	forkGenesisAlias = "genesis"
)

var forkOrder = []string{
	forkPhase0,
	"altair",
	"bellatrix",
	"capella",
	"deneb",
	"electra",
	"fulu",
	"gloas",
}

// SupportedForks returns the ordered set of consensus forks supported by the generator.
func SupportedForks() []string {
	out := make([]string, len(forkOrder))
	copy(out, forkOrder)
	return out
}

// ForkIndex returns a fork's activation order and accepts "genesis" as phase0.
func ForkIndex(name string) (int, bool) {
	name = strings.ToLower(name)
	if name == forkGenesisAlias {
		name = forkPhase0
	}
	for i, fork := range forkOrder {
		if fork == name {
			return i, true
		}
	}
	return 0, false
}

// ActiveForks returns all forks that must be configured up to the active fork.
func ActiveForks(active string) ([]string, error) {
	idx, ok := ForkIndex(active)
	if !ok {
		return nil, fmt.Errorf("unsupported fork %q", active)
	}
	out := make([]string, idx+1)
	copy(out, forkOrder[:idx+1])
	return out, nil
}

// ForkVersion returns a 4-byte fork version, using overrides before deriving one from chain ID.
func ForkVersion(chainID uint64, fork string, overrides map[string]string) string {
	if overrides != nil {
		if value, ok := overrides[strings.ToLower(fork)]; ok {
			return strings.ToLower(value)
		}
		if fork == forkPhase0 {
			if value, ok := overrides[forkGenesisAlias]; ok {
				return strings.ToLower(value)
			}
		}
	}

	idx, _ := ForkIndex(fork)
	value := (uint32(idx) << 24) | uint32(chainID&0x00ff_ffff)
	return fmt.Sprintf("0x%08x", value)
}

// ForkVersionField maps a fork name to the consensus config version field.
func ForkVersionField(fork string) string {
	if fork == forkPhase0 {
		return "GENESIS_FORK_VERSION"
	}
	return strings.ToUpper(fork) + "_FORK_VERSION"
}

// ForkEpochField maps a fork name to the consensus config epoch field.
func ForkEpochField(fork string) string {
	if fork == forkPhase0 {
		return ""
	}
	return strings.ToUpper(fork) + "_FORK_EPOCH"
}

func forkFieldName(name string) (string, bool) {
	name = strings.ToLower(name)
	if name == forkGenesisAlias {
		return "GENESIS_FORK_VERSION", true
	}
	if _, ok := ForkIndex(name); !ok {
		return "", false
	}
	return ForkVersionField(name), true
}
