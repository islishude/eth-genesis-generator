package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/islishude/bip39"
	appconfig "github.com/islishude/eth-genesis-generator/internal/config"
	"github.com/islishude/eth-genesis-generator/internal/consensus"
	"github.com/islishude/eth-genesis-generator/internal/execution"
)

// Generate writes all genesis artifacts into outDir and returns their manifest metadata.
func Generate(cfg *appconfig.Config, outDir string, now time.Time) (*Manifest, error) {
	if err := ensureDirs(outDir); err != nil {
		return nil, err
	}

	mnemonic := cfg.Consensus.Mnemonic
	generatedMnemonic := false
	if mnemonic == "" {
		var err error
		mnemonic, err = generateMnemonic()
		if err != nil {
			return nil, err
		}
		generatedMnemonic = true
	}

	elGenesis, err := execution.BuildGenesis(cfg)
	if err != nil {
		return nil, fmt.Errorf("build execution genesis: %w", err)
	}
	elBytes, err := execution.MarshalGenesis(elGenesis)
	if err != nil {
		return nil, fmt.Errorf("marshal execution genesis: %w", err)
	}
	elBytes = append(elBytes, '\n')

	clConfigBytes, err := consensus.RenderConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("render consensus config: %w", err)
	}
	mnemonicsBytes := consensus.RenderMnemonics(cfg, mnemonic)

	paths := map[string][]byte{
		"execution/genesis.json":    elBytes,
		"consensus/config.yaml":     []byte(clConfigBytes),
		"validators/mnemonics.yaml": []byte(mnemonicsBytes),
	}
	for relPath, data := range paths {
		// Start private because validators/mnemonics.yaml contains secret material.
		if err := os.WriteFile(filepath.Join(outDir, filepath.FromSlash(relPath)), data, 0o600); err != nil {
			return nil, fmt.Errorf("write %s: %w", relPath, err)
		}
	}
	// Public artifacts are relaxed after writing; the mnemonic file intentionally stays 0600.
	if err := os.Chmod(filepath.Join(outDir, "execution", "genesis.json"), 0o644); err != nil {
		return nil, err
	}
	if err := os.Chmod(filepath.Join(outDir, "consensus", "config.yaml"), 0o644); err != nil {
		return nil, err
	}

	genesisResult, err := consensus.BuildGenesis(
		elGenesis,
		filepath.Join(outDir, "consensus", "config.yaml"),
		filepath.Join(outDir, "validators", "mnemonics.yaml"),
		cfg.OutputJSONEnabled(),
	)
	if err != nil {
		return nil, err
	}

	artifactPaths := []string{
		"execution/genesis.json",
		"consensus/config.yaml",
		"consensus/genesis.ssz",
		"validators/mnemonics.yaml",
	}

	if err := os.WriteFile(filepath.Join(outDir, "consensus", "genesis.ssz"), genesisResult.SSZ, 0o644); err != nil {
		return nil, fmt.Errorf("write consensus/genesis.ssz: %w", err)
	}
	if cfg.OutputJSONEnabled() {
		jsonBytes := append(genesisResult.JSON, '\n')
		if err := os.WriteFile(filepath.Join(outDir, "consensus", "genesis.json"), jsonBytes, 0o644); err != nil {
			return nil, fmt.Errorf("write consensus/genesis.json: %w", err)
		}
		artifactPaths = append(artifactPaths, "consensus/genesis.json")
	}

	hashes, err := hashFiles(outDir, artifactPaths)
	if err != nil {
		return nil, err
	}

	manifest := &Manifest{
		NetworkName:          cfg.Network.Name,
		ChainID:              cfg.Network.ChainID,
		Fork:                 cfg.Consensus.Fork,
		GenesisTime:          cfg.Network.GenesisTime,
		ExecutionGenesisHash: elGenesis.ToBlock().Hash().Hex(),
		StateVersion:         genesisResult.StateVersion,
		ValidatorCount:       uint64(genesisResult.Validators),
		GeneratedMnemonic:    generatedMnemonic,
		GeneratedAtUnix:      now.Unix(),
		ArtifactHashes:       hashes,
	}

	if err := writeManifest(filepath.Join(outDir, "manifest.json"), manifest); err != nil {
		return nil, fmt.Errorf("write manifest.json: %w", err)
	}

	return manifest, nil
}

func generateMnemonic() (string, error) {
	return bip39.NewMnemonic(24, bip39.English)
}
