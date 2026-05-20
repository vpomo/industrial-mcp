package rest

import (
	"encoding/json"
	"net/http"
	"time"
)

type LicenseHandler struct {
	validator Validator
}

type Validator interface {
	IsEnabled() bool
	Validate() error
	GetHWID() string
	GetLicenseStatus() (bool, time.Time, []string)
	ValidateFeature(feature string) error
}

type LicenseStatusResponse struct {
	Enabled       bool      `json:"enabled"`
	Valid         bool      `json:"valid"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	Features      []string  `json:"features,omitempty"`
	DaysRemaining int       `json:"days_remaining,omitempty"`
}

type HWIDResponse struct {
	HardwareHash string `json:"hardware_hash"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewLicenseHandler(v Validator) *LicenseHandler {
	return &LicenseHandler{validator: v}
}

func (h *LicenseHandler) Status(w http.ResponseWriter, r *http.Request) {
	if !h.validator.IsEnabled() {
		resp := LicenseStatusResponse{Enabled: false, Valid: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	valid, expiresAt, features := h.validator.GetLicenseStatus()
	daysRemaining := 0
	if expiresAt.After(time.Now()) {
		daysRemaining = int(time.Until(expiresAt).Hours() / 24)
	}

	resp := LicenseStatusResponse{
		Enabled:       true,
		Valid:         valid,
		ExpiresAt:     expiresAt,
		Features:      features,
		DaysRemaining: daysRemaining,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *LicenseHandler) HWID(w http.ResponseWriter, r *http.Request) {
	hwid := h.validator.GetHWID()

	resp := HWIDResponse{HardwareHash: hwid}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *LicenseHandler) CheckFeature(w http.ResponseWriter, r *http.Request) {
	feature := r.URL.Query().Get("feature")
	if feature == "" {
		http.Error(w, "feature parameter required", http.StatusBadRequest)
		return
	}

	err := h.validator.ValidateFeature(feature)
	if err != nil {
		resp := ErrorResponse{Error: "feature_not_available", Message: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
}