package db

import "github.com/DAv10195/submit_commons/encryption"

var dbEncryption encryption.Encryption

func initDbEncryption(encryptionKeyFilePath string) error {
	if err := encryption.GenerateAesKeyFile(encryptionKeyFilePath); err != nil {
		return err
	}
	dbEncryption = &encryption.AesEncryption{KeyFilePath: encryptionKeyFilePath}
	return nil
}

func Decrypt(encryptedText string) (string, error) {
	decryptedText, err := dbEncryption.Decrypt(encryptedText)
	if err != nil {
		return "", err
	}
	return decryptedText, nil
}

func Encrypt(unEncryptedText string) (string, error) {
	encryptedText, err := dbEncryption.Encrypt(unEncryptedText)
	if err != nil {
		return "", err
	}
	return encryptedText, nil
}
