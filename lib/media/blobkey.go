package media

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// encodeBase62 encodes a byte slice to base62 string
func encodeBase62(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Convert bytes to big integer
	bigInt := new(big.Int).SetBytes(data)
	base := big.NewInt(62)
	zero := big.NewInt(0)

	// Convert to base62 by repeatedly dividing by 62
	var result []byte
	for bigInt.Cmp(zero) > 0 {
		remainder := new(big.Int)
		bigInt.DivMod(bigInt, base, remainder)
		// remainder is guaranteed to be < 62, so Int64() is safe
		result = append([]byte{base62Chars[remainder.Int64()]}, result...)
	}

	// Handle empty result (shouldn't happen with SHA256, but be safe)
	if len(result) == 0 {
		return "0"
	}

	return string(result)
}

// GenerateBlobKey generates a deterministic blob key from input string
// Uses SHA256 hash and encodes to base62 for URL-safe, compact representation
func GenerateBlobKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	encoded := encodeBase62(hash[:])
	return encoded
}

// GenerateSpectrogramBlobPath generates the full blob path for a spectrogram image
func GenerateSpectrogramBlobPath(paperTitle string) string {
	key := GenerateBlobKey(paperTitle)
	return fmt.Sprintf("media/papers/spectrogram/%s.webp", key)
}

