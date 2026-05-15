package license

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type LicenseValidator struct {
	crypto        *RSACrypto
	hardware      HardwareInfo
	enabled       bool
	checkInterval time.Duration
}

type LicenseData struct {
	HardwareHash string    `json:"hardware_hash"`
	ExpiresAt    time.Time `json:"expires_at"`
	Features     []string  `json:"features"`
	Signature    string    `json:"signature"`
}

func New(publicKeyPath string) (*LicenseValidator, error) {
	hw, err := GetHardwareInfo()
	if err != nil {
		return nil, err
	}

	var crypto *RSACrypto
	if publicKeyPath != "" {
		c, err := NewRSACrypto(publicKeyPath)
		if err != nil {
			return nil, err
		}
		crypto = c
	}

	return &LicenseValidator{
		crypto:        crypto,
		hardware:      hw,
		enabled:       os.Getenv("LICENSE_ENABLED") == "true",
		checkInterval: time.Hour,
	}, nil
}

func (v *LicenseValidator) IsEnabled() bool {
	return v.enabled
}

func (v *LicenseValidator) Validate() error {
	if !v.enabled {
		return nil
	}

	licenseData, err := v.loadLicenseFile()
	if err != nil {
		return err
	}

	if time.Now().After(licenseData.ExpiresAt) {
		return ErrLicenseExpired
	}

	if licenseData.HardwareHash != v.hardware.Hash() {
		return ErrHardwareMismatch
	}

	if v.crypto != nil && !v.crypto.Verify(v.hardware.Hash(), licenseData.Signature) {
		return ErrInvalidSignature
	}

	return nil
}

func (v *LicenseValidator) loadLicenseFile() (*LicenseData, error) {
	data, err := os.ReadFile("/app/license/license.dat")
	if err != nil {
		return nil, err
	}
	var license LicenseData
	if err := json.Unmarshal(data, &license); err != nil {
		return nil, err
	}
	return &license, nil
}

func (v *LicenseValidator) GetHardwareHash() string {
	return v.hardware.Hash()
}

var (
	ErrLicenseExpired   = errors.New("license expired")
	ErrHardwareMismatch = errors.New("hardware mismatch")
	ErrInvalidSignature = errors.New("invalid signature")
)