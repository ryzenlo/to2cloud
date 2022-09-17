package util

import (
	"fmt"
	"io/ioutil"
	"os"
)

const TempFilePrefix = "util"

type TempSSLCertFiles struct {
	tmpCertificate     *os.File
	tmpPrivateKey      *os.File
	certificateContent string
	privateKeyContent  string
}

func NewTempSSLCertFiles(cert, privateKey string) (*TempSSLCertFiles, error) {
	tf := &TempSSLCertFiles{
		certificateContent: cert,
		privateKeyContent:  privateKey,
	}
	var err error
	tf.tmpCertificate, err = ioutil.TempFile("/tmp", fmt.Sprintf("%s_ssl_cert_", TempFilePrefix))
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary ssl certificate file,%w", err)
	}
	//
	if _, err = tf.tmpCertificate.WriteString(cert); err != nil {
		return nil, fmt.Errorf("cannot write temporary ssh certificate file,%w", err)
	}
	if err = tf.tmpCertificate.Sync(); err != nil {
		return nil, fmt.Errorf("cannot write temporary ssh certificate file,%w", err)
	}
	//
	tf.tmpPrivateKey, err = ioutil.TempFile("/tmp", fmt.Sprintf("%s_ssl_private_key_", TempFilePrefix))
	if err != nil {
		return nil, fmt.Errorf("cannot create temporary ssl private file,%w", err)
	}
	//
	if _, err = tf.tmpPrivateKey.WriteString(privateKey); err != nil {
		return nil, fmt.Errorf("cannot write temporary ssl private file,%w", err)
	}
	if err = tf.tmpPrivateKey.Sync(); err != nil {
		return nil, fmt.Errorf("cannot write temporary ssl private file,%w", err)
	}
	return tf, nil
}

func (t *TempSSLCertFiles) GetCertificatePath() string {
	return t.tmpCertificate.Name()
}

func (t *TempSSLCertFiles) GetPrivateKeyPath() string {
	return t.tmpPrivateKey.Name()
}

func RemoveAllSSLTmpFiles(t *TempSSLCertFiles) {
	if t == nil {
		return
	}
	os.Remove(t.tmpCertificate.Name())
	os.Remove(t.tmpPrivateKey.Name())
}
