package x402

import (
	"context"
	"encoding/json"
	"errors"
	"os"
)

type Handler struct {
	enabled        bool
	paymentAddress string
	client         *Client
}

func NewHandler(enabled bool, paymentAddress string) *Handler {
	return &Handler{
		enabled:        enabled,
		paymentAddress: paymentAddress,
		client:         NewClient(paymentAddress),
	}
}

func NewHandlerFromEnv() *Handler {
	return &Handler{
		enabled:        os.Getenv("X402_ENABLED") == "true",
		paymentAddress: os.Getenv("X402_PAYMENT_ADDRESS"),
		client:         NewClient(os.Getenv("X402_PAYMENT_ADDRESS")),
	}
}

func (h *Handler) IsEnabled() bool {
	return h.enabled
}

func (h *Handler) ValidatePayment(ctx context.Context, params json.RawMessage) error {
	if !h.enabled {
		return nil
	}

	var paymentReq struct {
		Payment *PaymentRequest `json:"x402_payment,omitempty"`
	}

	if err := json.Unmarshal(params, &paymentReq); err != nil {
		return err
	}

	if paymentReq.Payment == nil {
		return ErrPaymentRequired
	}

	if paymentReq.Payment.Destination != h.paymentAddress {
		return ErrInvalidPaymentDestination
	}

	return nil
}

var (
	ErrPaymentRequired           = errors.New("payment required")
	ErrInvalidPaymentDestination = errors.New("invalid payment destination")
)