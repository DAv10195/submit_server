package fs

import (
	"fmt"
	"github.com/DAv10195/submit_commons/fsclient"
	"github.com/DAv10195/submit_server/db"
)

var wrapper *fileServerClientWrapper

type fileServerClientWrapper struct {
	client *fsclient.FileServerClient
}

func (w *fileServerClientWrapper) Decrypt(encryptedText string) (string, error) {
	return db.Decrypt(encryptedText)
}

func (w *fileServerClientWrapper) Encrypt(unencryptedText string) (string, error) {
	return db.Decrypt(unencryptedText)
}

func Init(fsHost string, fsPort int, fsUser string, fsPassword string) error {
	wrapper = &fileServerClientWrapper{}
	client, err := fsclient.NewFileServerClient(fmt.Sprintf("http://%s:%d", fsHost, fsPort), fsUser, fsPassword, logger, wrapper)
	if err != nil {
		return err
	}
	wrapper.client = client
	return nil
}

func GetClient() *fsclient.FileServerClient {
	if wrapper == nil || wrapper.client == nil {
		return nil
	}
	return wrapper.client
}
