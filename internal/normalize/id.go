package normalize

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateRowID(t Type, canonicalValue string) string {
	h := sha256.Sum256([]byte(string(t) + ":" + canonicalValue))
	return hex.EncodeToString(h[:16])
}
