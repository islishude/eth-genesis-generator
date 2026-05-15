package keystores

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/text/unicode/norm"
)

const (
	eip2335KDFPBKDF2       = "pbkdf2"
	eip2335PBKDF2Cost      = 1 << 18
	eip2335PBKDF2KeyLen    = 32
	eip2335PBKDF2PRF       = "hmac-sha256"
	eip2335CipherAES128CTR = "aes-128-ctr"
	eip2335ChecksumSHA256  = "sha256"
	eip2335SaltSize        = 32
	eip2335IVSize          = 16
)

type eip2335Crypto struct {
	KDF      eip2335KDF      `json:"kdf"`
	Checksum eip2335Checksum `json:"checksum"`
	Cipher   eip2335Cipher   `json:"cipher"`
}

type eip2335KDF struct {
	Function string           `json:"function"`
	Params   eip2335KDFParams `json:"params"`
	Message  string           `json:"message"`
}

type eip2335KDFParams struct {
	DKLen int    `json:"dklen"`
	C     int    `json:"c,omitempty"`
	N     int    `json:"n,omitempty"`
	P     int    `json:"p,omitempty"`
	R     int    `json:"r,omitempty"`
	PRF   string `json:"prf,omitempty"`
	Salt  string `json:"salt"`
}

type eip2335Checksum struct {
	Function string         `json:"function"`
	Params   map[string]any `json:"params"`
	Message  string         `json:"message"`
}

type eip2335Cipher struct {
	Function string              `json:"function"`
	Params   eip2335CipherParams `json:"params"`
	Message  string              `json:"message"`
}

type eip2335CipherParams struct {
	IV string `json:"iv"`
}

func encryptSecret(secret []byte, password string) (eip2335Crypto, error) {
	if secret == nil {
		return eip2335Crypto{}, fmt.Errorf("no secret")
	}

	salt := make([]byte, eip2335SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return eip2335Crypto{}, fmt.Errorf("generate salt: %w", err)
	}

	decryptionKey := deriveEIP2335DecryptionKey(salt, password)
	aesCipher, err := aes.NewCipher(decryptionKey[:16])
	if err != nil {
		return eip2335Crypto{}, fmt.Errorf("create aes cipher: %w", err)
	}

	iv := make([]byte, eip2335IVSize)
	if _, err := rand.Read(iv); err != nil {
		return eip2335Crypto{}, fmt.Errorf("generate initialization vector: %w", err)
	}

	cipherText := make([]byte, len(secret))
	cipher.NewCTR(aesCipher, iv).XORKeyStream(cipherText, secret)
	checksum := eip2335ChecksumMessage(decryptionKey, cipherText)

	return eip2335Crypto{
		KDF: eip2335KDF{
			Function: eip2335KDFPBKDF2,
			Params: eip2335KDFParams{
				DKLen: eip2335PBKDF2KeyLen,
				C:     eip2335PBKDF2Cost,
				PRF:   eip2335PBKDF2PRF,
				Salt:  hex.EncodeToString(salt),
			},
			Message: "",
		},
		Checksum: eip2335Checksum{
			Function: eip2335ChecksumSHA256,
			Params:   map[string]any{},
			Message:  hex.EncodeToString(checksum),
		},
		Cipher: eip2335Cipher{
			Function: eip2335CipherAES128CTR,
			Params: eip2335CipherParams{
				IV: hex.EncodeToString(iv),
			},
			Message: hex.EncodeToString(cipherText),
		},
	}, nil
}

func deriveEIP2335DecryptionKey(salt []byte, password string) []byte {
	return pbkdf2.Key(
		[]byte(normalizeKeystorePassword(password)),
		salt,
		eip2335PBKDF2Cost,
		eip2335PBKDF2KeyLen,
		sha256.New,
	)
}

func eip2335ChecksumMessage(decryptionKey []byte, cipherText []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(decryptionKey[16:32])
	_, _ = h.Write(cipherText)
	return h.Sum(nil)
}

func normalizeKeystorePassword(password string) string {
	var b strings.Builder
	for _, r := range norm.NFKD.String(password) {
		if r <= 0x1f || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
