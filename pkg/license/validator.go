package license

import (
	"errors"
	"os"
	"sync"
	"time"
)

var (
	ErrLicenseNotFound  = errors.New("license not found")
	ErrLicenseExpired   = errors.New("license expired")
	ErrHardwareMismatch = errors.New("hardware mismatch")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrLicenseCorrupted = errors.New("license corrupted")
	ErrFeatureMissing   = errors.New("feature missing")
)

type Validator struct {
	crypto        *RSACrypto
	hardware      HardwareInfo
	enabled       bool
	checkInterval time.Duration
	mu            sync.RWMutex
	licenseFile   *LicenseFile
	filePath      string
}

type ValidatorOption func(*Validator)

func WithCheckInterval(d time.Duration) ValidatorOption {
	return func(v *Validator) {
		v.checkInterval = d
	}
}

func New(publicKeyPEM []byte, filePath string, options ...ValidatorOption) (*Validator, error) {
	hw, err := GetHardwareInfo()
	if err != nil {
		return nil, err
	}

	var crypto *RSACrypto
	if len(publicKeyPEM) > 0 {
		c, err := NewRSACryptoFromPEM(publicKeyPEM)
		if err != nil {
			return nil, err
		}
		crypto = c
	}

	v := &Validator{
		crypto:        crypto,
		hardware:      hw,
		enabled:       os.Getenv("LICENSE_ENABLED") != "false",
		checkInterval: time.Hour,
		filePath:      filePath,
	}

	for _, opt := range options {
		opt(v)
	}

	return v, nil
}

func (v *Validator) IsEnabled() bool {
	return v.enabled
}

func (v *Validator) Validate() error {
	if !v.enabled {
		return nil
	}

	licenseFile := &LicenseFile{}
	if err := licenseFile.Load(v.filePath); err != nil {
		if os.IsNotExist(err) {
			return ErrLicenseNotFound
		}
		return ErrLicenseCorrupted
	}

	v.mu.Lock()
	v.licenseFile = licenseFile
	v.mu.Unlock()

	if err := v.validateLicenseFile(licenseFile); err != nil {
		return err
	}

	if v.crypto != nil {
		payload := licenseFile.Payload()
		if !v.crypto.Verify(payload, licenseFile.Signature) {
			return ErrInvalidSignature
		}
	}

	return nil
}

func (v *Validator) validateLicenseFile(lf *LicenseFile) error {
	if lf.IsExpired() {
		return ErrLicenseExpired
	}

	if lf.HardwareHash != v.hardware.Hash() {
		return ErrHardwareMismatch
	}

	return nil
}

func (v *Validator) ValidateFeature(feature string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.enabled {
		return nil
	}

	if v.licenseFile == nil {
		return ErrLicenseNotFound
	}

	if !v.licenseFile.HasFeature(feature) {
		return ErrFeatureMissing
	}

	return nil
}

func (v *Validator) GetHWID() string {
	return v.hardware.Hash()
}

func (v *Validator) GetLicenseStatus() (bool, time.Time, []string) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.licenseFile == nil {
		return false, time.Time{}, nil
	}

	return !v.licenseFile.IsExpired(), v.licenseFile.ExpiresAt, v.licenseFile.Features
}

func (v *Validator) StartPeriodicValidation(stopCh <-chan struct{}) {
	ticker := time.NewTicker(v.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := v.Validate(); err != nil {
				continue
			}
		case <-stopCh:
			return
		}
	}
}
