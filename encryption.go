package medego

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	aesBlockSize = aes.BlockSize
)

func (e *Encryption) Encrypt(text string) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aesBlockSize+len(plaintext))
	iv := ciphertext[:aesBlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aesBlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryption) Decrypt(text string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aesBlockSize {
		return "", fmt.Errorf("ciphertext is too short")
	}

	iv := ciphertext[:aesBlockSize]
	ciphertext = ciphertext[aesBlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
