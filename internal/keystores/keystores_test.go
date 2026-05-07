package keystores

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	bip39 "github.com/tyler-smith/go-bip39"
	e2util "github.com/wealdtech/go-eth2-util"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

func TestGenerateKeystoresFromMnemonics(t *testing.T) {
	const mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	const password = "test-password"

	outDir := t.TempDir()
	mnemonicsPath := filepath.Join(outDir, "mnemonics.yaml")
	if err := os.WriteFile(mnemonicsPath, []byte(`- mnemonic: "`+mnemonic+`"
  start: 2
  count: 2
`), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Generate(mnemonicsPath, outDir, password)
	if err != nil {
		t.Fatal(err)
	}
	if result.PasswordPath != PasswordRelPath {
		t.Fatalf("password path = %s", result.PasswordPath)
	}
	if len(result.KeystorePaths) != 2 {
		t.Fatalf("keystore count = %d", len(result.KeystorePaths))
	}

	passwordBytes, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(result.PasswordPath)))
	if err != nil {
		t.Fatal(err)
	}
	if string(passwordBytes) != password {
		t.Fatalf("password file = %q", string(passwordBytes))
	}
	assertFileMode(t, filepath.Join(outDir, filepath.FromSlash(result.PasswordPath)), 0o600)

	keystorePath := filepath.Join(outDir, filepath.FromSlash(result.KeystorePaths[0]))
	assertFileMode(t, keystorePath, 0o600)
	data, err := os.ReadFile(keystorePath)
	if err != nil {
		t.Fatal(err)
	}
	var keystore eip2335Keystore
	if err := json.Unmarshal(data, &keystore); err != nil {
		t.Fatal(err)
	}

	expectedPath := ValidatorKeyPath(2)
	if keystore.Path != expectedPath {
		t.Fatalf("path = %s", keystore.Path)
	}
	if keystore.Version != 4 {
		t.Fatalf("version = %d", keystore.Version)
	}
	if keystore.UUID == "" {
		t.Fatal("missing uuid")
	}

	seed := bip39.NewSeed(mnemonic, "")
	expectedKey, err := e2util.PrivateKeyFromSeedAndPath(seed, expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	if keystore.Pubkey != hex.EncodeToString(expectedKey.PublicKey().Marshal()) {
		t.Fatalf("pubkey = %s", keystore.Pubkey)
	}

	secret, err := keystorev4.New().Decrypt(keystore.Crypto, password)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(secret, expectedKey.Marshal()) {
		t.Fatal("decrypted secret does not match derived validator key")
	}
}

func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %o", path, got)
	}
}
