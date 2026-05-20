package license

import (
	"encoding/json"
	"os"
	"time"
)

type LicenseGenerator struct {
	crypto *RSACrypto
}

func NewLicenseGenerator(privateKeyPEM []byte) (*LicenseGenerator, error) {
	crypto, err := NewRSACryptoWithPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	return &LicenseGenerator{crypto: crypto}, nil
}

func (g *LicenseGenerator) Create(hardwareHash string, expiresAt time.Time, features []string, issuer string) (*LicenseFile, error) {
	issuedAt := time.Now()

	lf := &LicenseFile{
		Version:      1,
		HardwareHash: hardwareHash,
		IssuedAt:     issuedAt,
		ExpiresAt:    expiresAt,
		Features:     features,
		Issuer:       issuer,
	}

	payload := lf.Payload()
	signature, err := g.crypto.Sign(payload)
	if err != nil {
		return nil, err
	}
	lf.Signature = signature

	return lf, nil
}

func (g *LicenseGenerator) Save(lf *LicenseFile, path string) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}