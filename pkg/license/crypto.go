package license

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
)

var (
	ErrInvalidPEMBlock = errors.New("invalid PEM block")
	ErrInvalidKey      = errors.New("invalid key")
)

type RSACrypto struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewRSACryptoFromPEM(publicKeyPEM []byte) (*RSACrypto, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &RSACrypto{
		publicKey: pub.(*rsa.PublicKey),
	}, nil
}

func NewRSACryptoFromFile(path string) (*RSACrypto, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewRSACryptoFromPEM(keyData)
}

func NewRSACryptoWithPrivateKey(privateKeyPEM []byte) (*RSACrypto, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privPKCS1, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		priv = privPKCS1
	}

	rsaKey, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidKey
	}

	return &RSACrypto{
		privateKey: rsaKey,
		publicKey:  &rsaKey.PublicKey,
	}, nil
}

func (r *RSACrypto) Sign(data string) (string, error) {
	if r.privateKey == nil {
		return "", ErrInvalidKey
	}

	hashed := sha256.Sum256([]byte(data))

	signature, err := rsa.SignPKCS1v15(rand.Reader, r.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (r *RSACrypto) Verify(data, signatureBase64 string) bool {
	if r.publicKey == nil {
		return false
	}

	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}

	hashed := sha256.Sum256([]byte(data))

	return rsa.VerifyPKCS1v15(r.publicKey, crypto.SHA256, hashed[:], signature) == nil
}
