package cipher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// Encryption and decryption operations must use the same Aes instance
// ECB mode
type Aes struct {
	key       string
	aesCipher cipher.Block
	blockSize int
}

func NewAES(key string) *Aes {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil
	}
	return &Aes{
		key:       key,
		aesCipher: c,
		blockSize: len(key),
	}
}

var _ Cipher = (*Aes)(nil)

func (a *Aes) Encrypt(plainText []byte) ([]byte, error) {
	paddedText := padding(plainText, a.blockSize)
	out := make([]byte, len(paddedText))
	for times := 0; times < len(paddedText)/16; times++ {
		a.aesCipher.Encrypt(out[16*times:16*times+16], paddedText[16*times:16*times+16])
	}
	return out, nil
}

func (a *Aes) Decrypt(cipherText []byte) ([]byte, error) {
	if len(cipherText)%a.blockSize != 0 {
		panic("invalid cipher text length")
	}
	out := make([]byte, len(cipherText))
	for times := 0; times < len(cipherText)/16; times++ {
		a.aesCipher.Decrypt(out[16*times:16*times+16], cipherText[16*times:16*times+16])
	}
	return unpadding(out), nil
}

func padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func unpadding(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}
