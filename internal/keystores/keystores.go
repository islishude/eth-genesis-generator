package keystores

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	bip39 "github.com/tyler-smith/go-bip39"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	e2util "github.com/wealdtech/go-eth2-util"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	"gopkg.in/yaml.v3"
)

const (
	// DirRelPath is the artifact directory containing EIP-2335 validator keystores.
	DirRelPath = "validators/keystores"
	// PasswordRelPath is the generated password file used for all generated keystores.
	PasswordRelPath = "validators/keystores/password.txt"
)

var (
	initBLSOnce sync.Once
	initBLSErr  error
)

type mnemonicSource struct {
	Mnemonic string `yaml:"mnemonic"`
	Start    uint64 `yaml:"start"`
	Count    uint64 `yaml:"count"`
}

type eip2335Keystore struct {
	Crypto      map[string]any `json:"crypto"`
	Description string         `json:"description,omitempty"`
	Pubkey      string         `json:"pubkey,omitempty"`
	Path        string         `json:"path"`
	UUID        string         `json:"uuid"`
	Version     int            `json:"version"`
}

// Result summarizes generated keystore artifact paths relative to the output root.
type Result struct {
	KeystorePaths []string
	PasswordPath  string
}

// GeneratePassword returns a 32-byte random password encoded as hex ASCII.
func GeneratePassword() (string, error) {
	password := make([]byte, 32)
	if _, err := rand.Read(password); err != nil {
		return "", err
	}
	return hex.EncodeToString(password), nil
}

// Generate writes EIP-2335 validator keystores derived from mnemonicsPath.
func Generate(mnemonicsPath string, outDir string, password string) (*Result, error) {
	if password == "" {
		return nil, fmt.Errorf("keystore password must not be empty")
	}
	if err := initBLS(); err != nil {
		return nil, err
	}

	sources, err := loadMnemonicSources(mnemonicsPath)
	if err != nil {
		return nil, err
	}

	keystoreDir := filepath.Join(outDir, filepath.FromSlash(DirRelPath))
	if err := os.MkdirAll(keystoreDir, 0o700); err != nil {
		return nil, fmt.Errorf("create %s: %w", DirRelPath, err)
	}

	if err := os.WriteFile(filepath.Join(outDir, filepath.FromSlash(PasswordRelPath)), []byte(password), 0o600); err != nil {
		return nil, fmt.Errorf("write %s: %w", PasswordRelPath, err)
	}

	result := &Result{
		PasswordPath: PasswordRelPath,
	}
	encryptor := keystorev4.New()

	for _, source := range sources {
		seed, err := seedFromMnemonic(source.Mnemonic)
		if err != nil {
			return nil, err
		}
		for offset := uint64(0); offset < source.Count; offset++ {
			validatorIndex := source.Start + offset
			path := ValidatorKeyPath(validatorIndex)
			signingKey, err := e2util.PrivateKeyFromSeedAndPath(seed, path)
			if err != nil {
				return nil, fmt.Errorf("derive validator key %d: %w", validatorIndex, err)
			}

			crypto, err := encryptor.Encrypt(signingKey.Marshal(), password)
			if err != nil {
				return nil, fmt.Errorf("encrypt validator key %d: %w", validatorIndex, err)
			}
			id, err := uuid.NewRandom()
			if err != nil {
				return nil, fmt.Errorf("generate validator key uuid: %w", err)
			}

			keystore := eip2335Keystore{
				Crypto:      crypto,
				Description: fmt.Sprintf("eth-genesis-generator validator %d", validatorIndex),
				Pubkey:      hex.EncodeToString(signingKey.PublicKey().Marshal()),
				Path:        path,
				UUID:        id.String(),
				Version:     4,
			}
			data, err := json.MarshalIndent(keystore, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("marshal validator keystore %d: %w", validatorIndex, err)
			}
			data = append(data, '\n')

			relPath := filepath.ToSlash(filepath.Join(DirRelPath, fmt.Sprintf("keystore-%06d.json", len(result.KeystorePaths))))
			if err := os.WriteFile(filepath.Join(outDir, filepath.FromSlash(relPath)), data, 0o600); err != nil {
				return nil, fmt.Errorf("write %s: %w", relPath, err)
			}
			result.KeystorePaths = append(result.KeystorePaths, relPath)
		}
	}
	if len(result.KeystorePaths) == 0 {
		return nil, fmt.Errorf("no validator keystores generated from %s", mnemonicsPath)
	}

	return result, nil
}

// ValidatorKeyPath returns the ERC-2334 signing key path used by eth-beacon-genesis.
func ValidatorKeyPath(index uint64) string {
	return fmt.Sprintf("m/12381/3600/%d/0/0", index)
}

func initBLS() error {
	initBLSOnce.Do(func() {
		initBLSErr = e2types.InitBLS()
	})
	return initBLSErr
}

func loadMnemonicSources(path string) ([]mnemonicSource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var sources []mnemonicSource
	if err := yaml.Unmarshal(data, &sources); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no mnemonic entries in %s", path)
	}
	return sources, nil
}

func seedFromMnemonic(mnemonic string) ([]byte, error) {
	mnemonic = strings.TrimSpace(mnemonic)
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("mnemonic is not valid")
	}
	return bip39.NewSeed(mnemonic, ""), nil
}
