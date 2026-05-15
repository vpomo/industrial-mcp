package entity

import (
	"testing"
	"time"
)

func TestNewLicense(t *testing.T) {
	features := []string{"mqtt", "opcua"}
	license := NewLicense("lic-123", "hw-hash-abc", time.Now().Add(time.Hour), features)

	if license.ID() != "lic-123" {
		t.Errorf("expected ID 'lic-123', got %s", license.ID())
	}
	if license.HardwareHash() != "hw-hash-abc" {
		t.Errorf("expected HardwareHash 'hw-hash-abc', got %s", license.HardwareHash())
	}
	if !license.IsValid() {
		t.Error("expected IsValid to be true")
	}
}

func TestLicenseIsExpired(t *testing.T) {
	expiredLicense := NewLicense("expired", "hash", time.Now().Add(-time.Hour), nil)
	if !expiredLicense.IsExpired() {
		t.Error("expected expired license to be expired")
	}

	validLicense := NewLicense("valid", "hash", time.Now().Add(time.Hour), nil)
	if validLicense.IsExpired() {
		t.Error("expected valid license to not be expired")
	}
}

func TestLicenseHasFeature(t *testing.T) {
	features := []string{"mqtt", "opcua"}
	license := NewLicense("lic-1", "hash", time.Now().Add(time.Hour), features)

	if !license.HasFeature("mqtt") {
		t.Error("expected license to have mqtt feature")
	}
	if !license.HasFeature("opcua") {
		t.Error("expected license to have opcua feature")
	}
	if license.HasFeature("unknown") {
		t.Error("expected license to NOT have unknown feature")
	}
}

func TestLicenseInvalidate(t *testing.T) {
	license := NewLicense("lic-1", "hash", time.Now().Add(time.Hour), nil)
	license.Invalidate()
	if license.IsValid() {
		t.Error("expected license to be invalid after Invalidate()")
	}
}
