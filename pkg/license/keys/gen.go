package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

func Generate(keySize int) ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKey := &privateKey.PublicKey
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}

	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	})

	return privatePEM, publicPEM, nil
}

func Save(privatePEM, publicPEM []byte, dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "private.pem"), privatePEM, 0600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "public.pem"), publicPEM, 0644); err != nil {
		return err
	}
	return nil
}

func Load(dir string) ([]byte, []byte, error) {
	privatePEM, err := os.ReadFile(filepath.Join(dir, "private.pem"))
	if err != nil {
		return nil, nil, err
	}
	publicPEM, err := os.ReadFile(filepath.Join(dir, "public.pem"))
	if err != nil {
		return nil, nil, err
	}
	return privatePEM, publicPEM, nil
}
