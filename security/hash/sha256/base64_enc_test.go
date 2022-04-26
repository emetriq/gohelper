package sha256

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64EncHash(t *testing.T) {
	hash := Base64EncSha256([]byte("Hallo Welt"), []byte("0123456789abcdef"))
	assert.Equal(t, "YWyQmpWjRExz18r5eIppx/ZxI20QvXCNmZ+dJ5qtFGA=", hash)
}

func BenchmarkBase64EncHash(b *testing.B) {
	msg := []byte("Hallo Welt")
	key := []byte("0123456789abcdef")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Base64EncSha256(msg, key)
	}
}
