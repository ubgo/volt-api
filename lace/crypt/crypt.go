package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// https://gist.github.com/brettscott/2ac58ab7cb1c66e2b4a32d6c1c3908a7
// https://github.com/ent/ent/discussions/2364
// https://github.com/seamory/ent-encrypt-example/blob/main/internal/property/password.go

// Cipher key must be 32 chars long because block size is 16 bytes
// const CIPHER_KEY = "abcdefghijklmnopqrstuvwxyz012345"

var cipherKey string = "abcdefghijklmnopqrstuvwxyz012345"

func LoadKey(key string) {
	cipherKey = key
}

func Encrypt(unencrypted string) (string, error) {
	return EncryptWithKey(unencrypted, cipherKey)
}

// Encrypt encrypts plain text string into cipher text string
func EncryptWithKey(unencrypted string, cipherKey string) (string, error) {
	key := []byte(cipherKey)
	plainText := []byte(unencrypted)
	plainText, err := Pad(plainText, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf(`plainText: "%s" has error`, plainText)
	}
	if len(plainText)%aes.BlockSize != 0 {
		err := fmt.Errorf(`plainText: "%s" has the wrong block size`, plainText)
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	return fmt.Sprintf("%x", cipherText), nil
}

func Decrypt(encrypted string) (string, error) {
	return DecryptWithKey(encrypted, cipherKey)
}

// Decrypt decrypts cipher text string into plain text string
func DecryptWithKey(encrypted string, cipherKey string) (string, error) {
	key := []byte(cipherKey)
	cipherText, _ := hex.DecodeString(encrypted)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]

	// fmt.Printf("%X \n", iv)

	cipherText = cipherText[aes.BlockSize:]
	if len(cipherText)%aes.BlockSize != 0 {
		return "", errors.New("cipherText is not a multiple of the block size")
	}

	// fmt.Printf("%X \n", cipherText)

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	cipherText, err = Unpad(cipherText, aes.BlockSize)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", cipherText), nil
}

// https://github.com/mergermarket/go-pkcs7/blob/master/pkcs7.go
func Pad(buf []byte, size int) ([]byte, error) {
	bufLen := len(buf)
	padLen := size - bufLen%size
	padded := make([]byte, bufLen+padLen)
	copy(padded, buf)
	for i := 0; i < padLen; i++ {
		padded[bufLen+i] = byte(padLen)
	}
	return padded, nil
}

func Unpad(padded []byte, size int) ([]byte, error) {
	if len(padded)%size != 0 {
		return nil, errors.New("pkcs7: Padded value wasn't in correct size.")
	}

	bufLen := len(padded) - int(padded[len(padded)-1])

	// khanakia prevent panic
	if bufLen < 0 || bufLen > len(padded) {
		return nil, errors.New("pkcs7: Invalid padding on padded value.")
	}
	buf := make([]byte, bufLen)
	copy(buf, padded[:bufLen])
	return buf, nil
}
