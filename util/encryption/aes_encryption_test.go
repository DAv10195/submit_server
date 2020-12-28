package encryption

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	submitServer = "submit_server"
	aesKeyFileName = "submit_server_keystore"
)

func TestAesEncryptionDecryption(t *testing.T) {
	keyFilePath := filepath.Join(os.TempDir(), aesKeyFileName)
	if err := GenerateAesKeyFile(keyFilePath); err != nil {
		t.Fatal(err)
	}
	defer func(){
		if err := os.Remove(keyFilePath); err != nil {
			t.Fatal(err)
		}
	}()
	encryption := &AesEncryption{keyFilePath}
	encrypted, err := encryption.Encrypt(submitServer)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := encryption.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != submitServer {
		t.Fatalf("expected: \"%s\", but got: \"%s\"", submitServer, decrypted)
	}
}
