package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := "13800138000"

	encrypted, err := Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := Decrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDifferentNonce(t *testing.T) {
	key := make([]byte, 32)
	plaintext := "13800138000"

	e1, _ := Encrypt(plaintext, key)
	e2, _ := Encrypt(plaintext, key)
	assert.NotEqual(t, e1, e2)

	d1, _ := Decrypt(e1, key)
	d2, _ := Decrypt(e2, key)
	assert.Equal(t, d1, d2)
}

func TestEncryptInvalidKey(t *testing.T) {
	_, err := Encrypt("test", []byte("short"))
	assert.Error(t, err)
}

func TestDecryptInvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	_, err := Decrypt("not-base64!!!", key)
	assert.Error(t, err)
}

func TestHashSHA256(t *testing.T) {
	h1 := HashSHA256("13800138000")
	h2 := HashSHA256("13800138000")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64)

	h3 := HashSHA256("different")
	assert.NotEqual(t, h1, h3)
}

func TestMaskPhone(t *testing.T) {
	assert.Equal(t, "138****8000", MaskPhone("13800138000"))
	assert.Equal(t, "123", MaskPhone("123"))
	assert.Equal(t, "", MaskPhone(""))
}

func TestMaskIDCard(t *testing.T) {
	assert.Equal(t, "110101****1234", MaskIDCard("110101199001011234"))
	assert.Equal(t, "123", MaskIDCard("123"))
}
