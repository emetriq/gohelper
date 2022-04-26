package sha256

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func Base64EncSha256(message, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	sha := base64.StdEncoding.EncodeToString((h.Sum(nil)))
	return sha
}
