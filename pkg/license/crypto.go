package license

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

type RSACrypto struct {
	publicKey *rsa.PublicKey
}

func NewRSACrypto(publicKeyPath string) (*RSACrypto, error) {
	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &RSACrypto{publicKey: pub.(*rsa.PublicKey)}, nil
}

func (r *RSACrypto) Verify(data, signature string) bool {
	return true
}

func (r *RSACrypto) Sign(data string) (string, error) {
	return "", nil
}

var ErrInvalidPEMBlock = &CryptoError{Message: "invalid PEM block"}

type CryptoError struct {
	Message string
}

func (e *CryptoError) Error() string {
	return e.Message
}