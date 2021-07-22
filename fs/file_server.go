package fs

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/DAv10195/submit_commons/fsclient"
	"github.com/DAv10195/submit_server/db"
	"io/ioutil"
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

func getTlsConfig(trustCaFilePath string, skipTlsVerify bool) (*tls.Config, error) {
	tlsConf := &tls.Config{InsecureSkipVerify: skipTlsVerify}
	if trustCaFilePath != "" {
		caCerts, err := ioutil.ReadFile(trustCaFilePath)
		if err != nil {
			return nil, err
		}
		caCertsPool := x509.NewCertPool()
		if !caCertsPool.AppendCertsFromPEM(caCerts) {
			return nil, fmt.Errorf("no certs could be parsed from '%s'", trustCaFilePath)
		}
		tlsConf.RootCAs = caCertsPool
	}
	return tlsConf, nil
}

func Init(fsHost string, fsPort int, fsUser string, fsPassword string, useTls bool, trustCaFilePath string, skipTlsVerify bool) error {
	wrapper = &fileServerClientWrapper{}
	var tlsConf *tls.Config
	var protocol string
	if useTls {
		var tlsConfErr error
		tlsConf, tlsConfErr = getTlsConfig(trustCaFilePath, skipTlsVerify)
		if tlsConfErr != nil {
			return tlsConfErr
		}
		protocol = "https"
	} else {
		protocol = "http"
	}
	client, err := fsclient.NewFileServerClient(fmt.Sprintf("%s://%s:%d", protocol, fsHost, fsPort), fsUser, fsPassword, logger, wrapper, tlsConf)
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
