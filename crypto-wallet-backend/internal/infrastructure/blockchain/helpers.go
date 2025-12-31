package blockchain

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"time"
)

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func randomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("blockchain: length must be positive")
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

func encodeBase58(data []byte) string {
	intData := new(big.Int).SetBytes(data)
	zero := big.NewInt(0)
	base := big.NewInt(58)
	mod := new(big.Int)

	var result []byte
	for intData.Cmp(zero) > 0 {
		intData.DivMod(intData, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	for _, b := range data {
		if b == 0x00 {
			result = append(result, base58Alphabet[0])
		} else {
			break
		}
	}

	// Reverse result in-place.
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func encodeBase32Lower(data []byte) string {
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	return strings.ToLower(encoding.EncodeToString(data))
}

func encodeBase32Upper(data []byte) string {
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	return strings.ToUpper(encoding.EncodeToString(data))
}

func encodeHexLower(data []byte) string {
	return strings.ToLower(hex.EncodeToString(data))
}

func randomPublicKeyString() (string, error) {
	bytes, err := randomBytes(33)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func synthBalance(address string, confirmations int) *Balance {
	return &Balance{
		Address:       address,
		Balance:       "0",
		Confirmations: confirmations,
		LastUpdated:   time.Now().UTC(),
	}
}
