package config

import (
	"fmt"
	"strings"
)

// ParseExecutionContractProfiles converts a comma-separated CLI value into a
// normalized execution contract profile list.
func ParseExecutionContractProfiles(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("execution.contracts must not be empty; supported profiles: %s", SupportedExecutionContractProfiles())
	}
	return NormalizeExecutionContractProfiles(strings.Split(value, ","))
}

// NormalizeExecutionContractProfiles validates and canonicalizes execution
// contract predeploy profile names.
func NormalizeExecutionContractProfiles(profiles []string) ([]string, error) {
	normalized := make([]string, 0, len(profiles))
	seen := make(map[string]struct{}, len(profiles))
	for _, raw := range profiles {
		profile := strings.ToLower(strings.TrimSpace(raw))
		if profile == "" {
			return nil, fmt.Errorf("execution.contracts contains an empty profile; supported profiles: %s", SupportedExecutionContractProfiles())
		}
		if !isSupportedExecutionContractProfile(profile) {
			return nil, fmt.Errorf("execution.contracts has unsupported profile %q; supported profiles: %s", raw, SupportedExecutionContractProfiles())
		}
		if _, exists := seen[profile]; exists {
			return nil, fmt.Errorf("execution.contracts has duplicate profile %q", profile)
		}
		seen[profile] = struct{}{}
		normalized = append(normalized, profile)
	}

	if len(normalized) == 0 {
		return nil, nil
	}
	if _, hasNone := seen[ExecutionContractsProfileNone]; hasNone {
		if len(normalized) > 1 {
			return nil, fmt.Errorf("execution.contracts profile %q cannot be combined with other profiles", ExecutionContractsProfileNone)
		}
		return nil, nil
	}
	if _, hasAll := seen[ExecutionContractsProfileAll]; hasAll && len(normalized) > 1 {
		return nil, fmt.Errorf("execution.contracts profile %q cannot be combined with other profiles", ExecutionContractsProfileAll)
	}
	return normalized, nil
}

// SupportedExecutionContractProfiles returns the supported profile names for
// CLI help and validation errors.
func SupportedExecutionContractProfiles() string {
	return strings.Join([]string{
		ExecutionContractsProfileSystem,
		ExecutionContractsProfileUtils,
		ExecutionContractsProfileAll,
		ExecutionContractsProfileNone,
	}, ", ")
}

func isSupportedExecutionContractProfile(profile string) bool {
	switch profile {
	case ExecutionContractsProfileSystem,
		ExecutionContractsProfileUtils,
		ExecutionContractsProfileAll,
		ExecutionContractsProfileNone:
		return true
	default:
		return false
	}
}
