package aes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESSuccess(t *testing.T) {
	msg := []byte("Hallo Welt")
	key := []byte("0123456789abcdef")
	encrypted, err := Encrypt(msg, key)
	assert.Nil(t, err)
	decrypted, err := Decrypt(encrypted, key)
	assert.Nil(t, err)
	assert.Equal(t, msg, decrypted)
}

func TestAESWrongKey(t *testing.T) {
	msg := []byte("Hallo Welt")
	key := []byte("0123456789abcdef")
	encrypted, err := Encrypt(msg, []byte("23456789abcdef01"))
	assert.Nil(t, err)
	decrypted, err := Decrypt(encrypted, key)
	assert.NotNil(t, err)
	assert.Nil(t, decrypted)
}

func BenchmarkAESEncrypt(b *testing.B) {
	msg := []byte("Hallo Welt")
	key := []byte("0123456789abcdef")
	for i := 0; i < b.N; i++ {
		Encrypt(msg, key)
	}
}

func BenchmarkAESDecrypt(b *testing.B) {
	msg := []byte("Hallo Welt")
	key := []byte("0123456789abcdef")
	result, _ := Encrypt(msg, key)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decrypt(result, key)
	}
}
