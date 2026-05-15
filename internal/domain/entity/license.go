package entity

import (
	"time"
)

type License struct {
	id           string
	hardwareHash string
	expiresAt    time.Time
	features     []string
	isValid      bool
}

func NewLicense(id, hardwareHash string, expiresAt time.Time, features []string) *License {
	return &License{
		id:           id,
		hardwareHash: hardwareHash,
		expiresAt:    expiresAt,
		features:     features,
		isValid:      true,
	}
}

func (l *License) ID() string    { return l.id }
func (l *License) HardwareHash() string { return l.hardwareHash }
func (l *License) ExpiresAt() time.Time { return l.expiresAt }
func (l *License) Features() []string { return l.features }
func (l *License) IsValid() bool { return l.isValid }

func (l *License) IsExpired() bool {
	return time.Now().After(l.expiresAt)
}

func (l *License) HasFeature(feature string) bool {
	for _, f := range l.features {
		if f == feature {
			return true
		}
	}
	return false
}

func (l *License) Invalidate() {
	l.isValid = false
}