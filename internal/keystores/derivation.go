package keystores

import (
	"fmt"
	"strconv"
	"strings"

	blst "github.com/supranational/blst/bindings/go"
)

const eip2333MinSeedLen = 32

type validatorKey struct {
	secret []byte
	pubkey []byte
}

func deriveValidatorKey(seed []byte, path string) (*validatorKey, error) {
	secretKey, err := deriveSecretKeyFromSeedAndPath(seed, path)
	if err != nil {
		return nil, err
	}
	defer secretKey.Zeroize()

	pubkey := new(blst.P1Affine).From(secretKey).Compress()
	return &validatorKey{
		secret: append([]byte(nil), secretKey.Serialize()...),
		pubkey: append([]byte(nil), pubkey...),
	}, nil
}

func deriveSecretKeyFromSeedAndPath(seed []byte, path string) (*blst.SecretKey, error) {
	if path == "" {
		return nil, fmt.Errorf("no path")
	}
	if len(seed) < eip2333MinSeedLen {
		return nil, fmt.Errorf("seed must be at least %d bytes", eip2333MinSeedLen)
	}

	pathBits := strings.Split(path, "/")
	var secretKey *blst.SecretKey
	for i, pathBit := range pathBits {
		switch pathBit {
		case "":
			return nil, fmt.Errorf("no entry at path component %d", i)
		case "m":
			if i != 0 {
				return nil, fmt.Errorf("invalid master at path component %d", i)
			}
			secretKey = blst.DeriveMasterEip2333(seed)
			if secretKey == nil || !secretKey.Valid() {
				return nil, fmt.Errorf("derive master key")
			}
		default:
			if i == 0 {
				return nil, fmt.Errorf("not master at path component %d", i)
			}
			if secretKey == nil {
				return nil, fmt.Errorf("missing master key before path component %d", i)
			}
			index, err := strconv.ParseUint(pathBit, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid index %q at path component %d", pathBit, i)
			}
			secretKey = secretKey.DeriveChildEip2333(uint32(index))
			if secretKey == nil || !secretKey.Valid() {
				return nil, fmt.Errorf("derive child key at path component %d", i)
			}
		}
	}

	return secretKey, nil
}
