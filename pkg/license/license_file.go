package license

import (
	"encoding/json"
	"os"
	"time"
)

type LicenseFile struct {
	Version      int       `json:"version"`
	HardwareHash string    `json:"hardware_hash"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Features     []string  `json:"features"`
	Signature    string    `json:"signature"`
	Issuer       string    `json:"issuer"`
}

func (f *LicenseFile) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, f)
}

func (f *LicenseFile) Save(path string) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (f *LicenseFile) IsExpired() bool {
	return time.Now().After(f.ExpiresAt)
}

func (f *LicenseFile) HasFeature(feature string) bool {
	for _, f := range f.Features {
		if f == feature {
			return true
		}
	}
	return false
}

func (f *LicenseFile) Payload() string {
	featuresJSON, _ := json.Marshal(f.Features)
	return f.HardwareHash + "|" + f.ExpiresAt.Format(time.RFC3339) + "|" + string(featuresJSON)
}