package encryption

type Encryption interface {
	// return a decrypted version of the given encrypted string
	Decrypt(encryptedText string) (string, error)
	// return an encrypted version of the given unencrypted string
	Encrypt(unencryptedText string) (string, error)
}
