package keystores

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	bip39 "github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
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
	expectedKey, err := deriveValidatorKey(seed, expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	if keystore.Pubkey != hex.EncodeToString(expectedKey.pubkey) {
		t.Fatalf("pubkey = %s", keystore.Pubkey)
	}

	assertDefaultPBKDF2Crypto(t, keystore.Crypto)
	secret := decryptSecret(t, keystore.Crypto, password)
	if !bytes.Equal(secret, expectedKey.secret) {
		t.Fatal("decrypted secret does not match derived validator key")
	}
}

func TestEIP2335OfficialTestVectors(t *testing.T) {
	const password = "testpassword🔑"
	const encodedPassword = "7465737470617373776f7264f09f9491"
	const secret = "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"

	if got := hex.EncodeToString([]byte(normalizeKeystorePassword(password))); got != encodedPassword {
		t.Fatalf("encoded password = %s", got)
	}

	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Scrypt",
			input: `{
				"crypto": {
					"kdf": {
						"function": "scrypt",
						"params": {
							"dklen": 32,
							"n": 262144,
							"p": 1,
							"r": 8,
							"salt": "d4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"
						},
						"message": ""
					},
					"checksum": {
						"function": "sha256",
						"params": {},
						"message": "d2217fe5f3e9a1e34581ef8a78f7c9928e436d36dacc5e846690a5581e8ea484"
					},
					"cipher": {
						"function": "aes-128-ctr",
						"params": {
							"iv": "264daa3f303d7259501c93d997d84fe6"
						},
						"message": "06ae90d55fe0a6e9c5c3bc5b170827b2e5cce3929ed3f116c2811e6366dfe20f"
					}
				},
				"description": "This is a test keystore that uses scrypt to secure the secret.",
				"pubkey": "9612d7a727c9d0a22e185a1c768478dfe919cada9266988cb32359c11f2b7b27f4ae4040902382ae2910c15e2b420d07",
				"path": "m/12381/60/3141592653/589793238",
				"uuid": "1d85ae20-35c5-4611-98e8-aa14a633906f",
				"version": 4
			}`,
		},
		{
			name: "PBKDF2",
			input: `{
				"crypto": {
					"kdf": {
						"function": "pbkdf2",
						"params": {
							"dklen": 32,
							"c": 262144,
							"prf": "hmac-sha256",
							"salt": "d4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"
						},
						"message": ""
					},
					"checksum": {
						"function": "sha256",
						"params": {},
						"message": "8a9f5d9912ed7e75ea794bc5a89bca5f193721d30868ade6f73043c6ea6febf1"
					},
					"cipher": {
						"function": "aes-128-ctr",
						"params": {
							"iv": "264daa3f303d7259501c93d997d84fe6"
						},
						"message": "cee03fde2af33149775b7223e7845e4fb2c8ae1792e5f99fe9ecf474cc8c16ad"
					}
				},
				"description": "This is a test keystore that uses PBKDF2 to secure the secret.",
				"pubkey": "9612d7a727c9d0a22e185a1c768478dfe919cada9266988cb32359c11f2b7b27f4ae4040902382ae2910c15e2b420d07",
				"path": "m/12381/60/0/0",
				"uuid": "64625def-3331-4eea-ab6f-782f3ed16a83",
				"version": 4
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var keystore eip2335Keystore
			if err := json.Unmarshal([]byte(test.input), &keystore); err != nil {
				t.Fatal(err)
			}
			if keystore.Version != 4 {
				t.Fatalf("version = %d", keystore.Version)
			}
			if keystore.Pubkey != "9612d7a727c9d0a22e185a1c768478dfe919cada9266988cb32359c11f2b7b27f4ae4040902382ae2910c15e2b420d07" {
				t.Fatalf("pubkey = %s", keystore.Pubkey)
			}
			if got := hex.EncodeToString(decryptSecret(t, keystore.Crypto, password)); got != secret {
				t.Fatalf("secret = %s", got)
			}
		})
	}
}

func TestDeriveValidatorKeyEIP2333Vectors(t *testing.T) {
	tests := []struct {
		name            string
		seedHex         string
		childPath       string
		masterSecretDec string
		childSecretDec  string
	}{
		{
			name:            "TestCase0",
			seedHex:         "c55257c360c07c72029aebc1b53c05ed0362ada38ead3e3e9efa3708e53495531f09a6987599d18264c1e1c92f2cf141630c7a3c4ab7c81b2f001698e7463b04",
			childPath:       "m/0",
			masterSecretDec: "6083874454709270928345386274498605044986640685124978867557563392430687146096",
			childSecretDec:  "20397789859736650942317412262472558107875392172444076792671091975210932703118",
		},
		{
			name:            "TestCase1",
			seedHex:         "3141592653589793238462643383279502884197169399375105820974944592",
			childPath:       "m/3141592653",
			masterSecretDec: "29757020647961307431480504535336562678282505419141012933316116377660817309383",
			childSecretDec:  "25457201688850691947727629385191704516744796114925897962676248250929345014287",
		},
		{
			name:            "TestCase2",
			seedHex:         "0099FF991111002299DD7744EE3355BBDD8844115566CC55663355668888CC00",
			childPath:       "m/4294967295",
			masterSecretDec: "27580842291869792442942448775674722299803720648445448686099262467207037398656",
			childSecretDec:  "29358610794459428860402234341874281240803786294062035874021252734817515685787",
		},
		{
			name:            "TestCase3",
			seedHex:         "d4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
			childPath:       "m/42",
			masterSecretDec: "19022158461524446591288038168518313374041767046816487870552872741050760015818",
			childSecretDec:  "31372231650479070279774297061823572166496564838472787488249775572789064611981",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seed := mustDecodeHex(t, test.seedHex)
			masterKey, err := deriveSecretKeyFromSeedAndPath(seed, "m")
			if err != nil {
				t.Fatal(err)
			}
			defer masterKey.Zeroize()
			if !bytes.Equal(masterKey.Serialize(), mustSecretDecimal(t, test.masterSecretDec)) {
				t.Fatalf("master secret = 0x%x", masterKey.Serialize())
			}

			key, err := deriveValidatorKey(seed, test.childPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(key.secret, mustSecretDecimal(t, test.childSecretDec)) {
				t.Fatalf("secret = 0x%x", key.secret)
			}
			if len(key.pubkey) != 48 {
				t.Fatalf("pubkey length = %d", len(key.pubkey))
			}
		})
	}
}

func TestDeriveValidatorKeyKnownBeaconGenesisPubkey(t *testing.T) {
	const mnemonic = "rare observe fox place unfold bargain cannon direct title sorry rabbit juice body autumn quality decrease mixture transfer crisp unveil path depend brick scissors"

	key, err := deriveValidatorKey(bip39.NewSeed(mnemonic, ""), ValidatorKeyPath(0))
	if err != nil {
		t.Fatal(err)
	}

	const want = "a72ce460a5ab6bea347e59b17ee349bebf6adfa0a240993ed70a5be0da9638b6e2dc7bbdd19e24a8292c1c7b30f23c9e"
	if got := hex.EncodeToString(key.pubkey); got != want {
		t.Fatalf("pubkey = %s", got)
	}
}

func TestDeriveValidatorKeyRejectsInvalidInputs(t *testing.T) {
	seed := mustDecodeHex(t, "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	tests := []string{
		"",
		"m/m/12381",
		"1/m/12381",
		"m/12381//0",
		"m/4294967296",
		"m/-1",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			if _, err := deriveValidatorKey(seed, path); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	if _, err := deriveValidatorKey(seed[:31], "m/0"); err == nil {
		t.Fatal("expected short seed error")
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

func assertDefaultPBKDF2Crypto(t *testing.T, crypto eip2335Crypto) {
	t.Helper()
	if crypto.KDF.Function != eip2335KDFPBKDF2 {
		t.Fatalf("kdf function = %s", crypto.KDF.Function)
	}
	if crypto.KDF.Params.DKLen != eip2335PBKDF2KeyLen {
		t.Fatalf("kdf dklen = %d", crypto.KDF.Params.DKLen)
	}
	if crypto.KDF.Params.C != eip2335PBKDF2Cost {
		t.Fatalf("kdf cost = %d", crypto.KDF.Params.C)
	}
	if crypto.KDF.Params.PRF != eip2335PBKDF2PRF {
		t.Fatalf("kdf prf = %s", crypto.KDF.Params.PRF)
	}
	if crypto.Checksum.Function != eip2335ChecksumSHA256 {
		t.Fatalf("checksum function = %s", crypto.Checksum.Function)
	}
	if crypto.Cipher.Function != eip2335CipherAES128CTR {
		t.Fatalf("cipher function = %s", crypto.Cipher.Function)
	}
}

func decryptSecret(t *testing.T, crypto eip2335Crypto, password string) []byte {
	t.Helper()
	if crypto.Checksum.Function != eip2335ChecksumSHA256 {
		t.Fatalf("checksum function = %s", crypto.Checksum.Function)
	}
	if crypto.Cipher.Function != eip2335CipherAES128CTR {
		t.Fatalf("cipher function = %s", crypto.Cipher.Function)
	}
	salt := mustDecodeHex(t, crypto.KDF.Params.Salt)
	decryptionKey := deriveTestDecryptionKey(t, crypto.KDF, salt, password)
	cipherText := mustDecodeHex(t, crypto.Cipher.Message)
	hash := sha256.New()
	_, _ = hash.Write(decryptionKey[16:32])
	_, _ = hash.Write(cipherText)
	if !bytes.Equal(hash.Sum(nil), mustDecodeHex(t, crypto.Checksum.Message)) {
		t.Fatal("invalid checksum")
	}

	block, err := aes.NewCipher(decryptionKey[:16])
	if err != nil {
		t.Fatal(err)
	}
	iv := mustDecodeHex(t, crypto.Cipher.Params.IV)
	secret := make([]byte, len(cipherText))
	cipher.NewCTR(block, iv).XORKeyStream(secret, cipherText)
	return secret
}

func deriveTestDecryptionKey(t *testing.T, kdf eip2335KDF, salt []byte, password string) []byte {
	t.Helper()
	switch kdf.Function {
	case eip2335KDFPBKDF2:
		if kdf.Params.PRF != eip2335PBKDF2PRF {
			t.Fatalf("kdf prf = %s", kdf.Params.PRF)
		}
		return pbkdf2.Key(
			[]byte(normalizeKeystorePassword(password)),
			salt,
			kdf.Params.C,
			kdf.Params.DKLen,
			sha256.New,
		)
	case "scrypt":
		key, err := scrypt.Key(
			[]byte(normalizeKeystorePassword(password)),
			salt,
			kdf.Params.N,
			kdf.Params.R,
			kdf.Params.P,
			kdf.Params.DKLen,
		)
		if err != nil {
			t.Fatal(err)
		}
		return key
	default:
		t.Fatalf("unsupported kdf function %s", kdf.Function)
		return nil
	}
}

func mustDecodeHex(t *testing.T, input string) []byte {
	t.Helper()
	data, err := hex.DecodeString(input)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mustSecretDecimal(t *testing.T, input string) []byte {
	t.Helper()
	value, ok := new(big.Int).SetString(input, 10)
	if !ok {
		t.Fatalf("invalid decimal %q", input)
	}
	out := make([]byte, 32)
	bytes := value.Bytes()
	copy(out[len(out)-len(bytes):], bytes)
	return out
}
