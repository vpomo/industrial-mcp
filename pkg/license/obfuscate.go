package license

import (
	"encoding/hex"
	"strings"
)

const obfuscationKey = "imatic_license_2024_secure_key_factor"

type ObfuscatedKey struct {
	parts []string
}

func ObfuscatePEMKey(pemData []byte) ObfuscatedKey {
	hexStr := hex.EncodeToString(pemData)
	partLen := len(hexStr) / 4

	parts := make([]string, 4)
	for i := 0; i < 4; i++ {
		start := i * partLen
		end := start + partLen
		if i == 3 {
			end = len(hexStr)
		}
		parts[i] = xorString(hexStr[start:end], obfuscationKey)
	}

	return ObfuscatedKey{parts: parts}
}

func (o ObfuscatedKey) Assemble() []byte {
	var builder strings.Builder
	for _, part := range o.parts {
		builder.WriteString(xorString(part, obfuscationKey))
	}
	data, _ := hex.DecodeString(builder.String())
	return data
}

func xorString(s, key string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[i] ^ key[i%len(key)]
	}
	return string(result)
}

func (o ObfuscatedKey) Parts() []string {
	return o.parts
}

func NewObfuscatedKeyFromParts(parts []string) ObfuscatedKey {
	return ObfuscatedKey{parts: parts}
}