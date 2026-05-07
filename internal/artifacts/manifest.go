package artifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest summarizes generated artifacts and records hashes for reproducibility checks.
type Manifest struct {
	// NetworkName is copied from the user config for quick artifact identification.
	NetworkName string `json:"network_name"`
	// ChainID is the execution and deposit chain ID.
	ChainID uint64 `json:"chain_id"`
	// Fork is the consensus fork configured active at genesis.
	Fork string `json:"fork"`
	// GenesisTime is the Unix timestamp used by both layers.
	GenesisTime uint64 `json:"genesis_time"`
	// ExecutionGenesisHash is the geth block hash for execution/genesis.json.
	ExecutionGenesisHash string `json:"execution_genesis_hash"`
	// StateVersion is the beacon state version reported by eth-beacon-genesis.
	StateVersion string `json:"state_version"`
	// ValidatorCount is the number of validators in consensus/genesis.ssz.
	ValidatorCount uint64 `json:"validator_count"`
	// ValidatorKeystoreCount is the number of generated EIP-2335 validator keystores.
	ValidatorKeystoreCount uint64 `json:"validator_keystore_count"`
	// GeneratedMnemonic reports whether the CLI generated validator secret material.
	GeneratedMnemonic bool `json:"generated_mnemonic"`
	// GeneratedAtUnix is the wall-clock time when artifacts were written.
	GeneratedAtUnix int64 `json:"generated_at_unix"`
	// ArtifactHashes maps relative artifact paths to SHA-256 hex digests.
	ArtifactHashes map[string]string `json:"artifact_hashes"`
}

func writeManifest(path string, manifest *Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func hashFiles(root string, relPaths []string) (map[string]string, error) {
	hashes := make(map[string]string, len(relPaths))
	for _, relPath := range relPaths {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		hashes[relPath] = hex.EncodeToString(sum[:])
	}
	return hashes, nil
}

func ensureDirs(root string) error {
	for _, dir := range []string{"execution", "consensus", "validators"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
