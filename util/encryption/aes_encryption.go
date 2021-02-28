package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	aesKeyLength = 32 // 256 bits
	aesKeyFilePerms = 0600
)

// AES encryption
type AesEncryption struct {
	KeyFilePath	string
}

// decrypt the given base64 encoded string
func (e *AesEncryption) Decrypt(encryptedText string) (string, error) {
	decryptedText, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("error decrypting text: %v", err)
	}
	key, err := e.getKeyFromFile()
	if err != nil {
		return "", fmt.Errorf("error decrypting text: %v", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error decrypting text: %v", err)
	}
	if len(decryptedText) < aes.BlockSize {
		return "", fmt.Errorf("error decrypting text: number of bytes in the text given for decryption (\"%s\") is less than the AES block size (%d)", encryptedText, aes.BlockSize)
	}
	iv, decryptedText := decryptedText[ : aes.BlockSize], decryptedText[aes.BlockSize : ]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(decryptedText, decryptedText)
	return string(decryptedText), nil
}

// encrypt the given string and return a base64 encoded string of the encrypted value
func (e *AesEncryption) Encrypt(unencryptedText string) (string, error) {
	key, err := e.getKeyFromFile()
	if err != nil {
		return "", fmt.Errorf("error encrypting text: %v", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error encrypting text: %v", err)
	}
	encryptedText := make([]byte, aes.BlockSize + len([]byte(unencryptedText)))
	iv := encryptedText[ : aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("error encrypting text: %v", err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(encryptedText[aes.BlockSize : ], []byte(unencryptedText))
	return base64.StdEncoding.EncodeToString(encryptedText), nil
}

func (e *AesEncryption) getKeyFromFile() ([]byte, error) {
	keyFromFile, err := ioutil.ReadFile(e.KeyFilePath)
	if err != nil {
		return nil, err
	}
	decodedKey := make([]byte, aesKeyLength)
	numDecodedBytes, err := base64.StdEncoding.Decode(decodedKey, keyFromFile)
	if numDecodedBytes != aesKeyLength {
		return nil, fmt.Errorf("number of bytes in key file (%s) is not as expected (%d)", e.KeyFilePath, aesKeyLength)
	}
	return decodedKey, nil
}

func GenerateAesKeyFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		key := make([]byte, aesKeyLength)
		if _, err := rand.Read(key); err != nil {
			return err
		}
		encodedKey := make([]byte, base64.StdEncoding.EncodedLen(aesKeyLength))
		base64.StdEncoding.Encode(encodedKey, key)
		if err := ioutil.WriteFile(path, encodedKey, aesKeyFilePerms); err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}
