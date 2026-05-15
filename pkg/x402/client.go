package x402

import (
	"context"
	"encoding/json"
	"net/http"
)

type Client struct {
	paymentAddress string
	httpClient     *http.Client
}

type PaymentRequest struct {
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Destination string `json:"destination"`
	Memo        string `json:"memo"`
}

type PaymentResponse struct {
	Paid       bool   `json:"paid"`
	PaymentURL string `json:"payment_url,omitempty"`
}

func NewClient(paymentAddress string) *Client {
	return &Client{
		paymentAddress: paymentAddress,
		httpClient:     &http.Client{},
	}
}

func (c *Client) CreatePayment(ctx context.Context, amount, currency, memo string) (*PaymentResponse, error) {
	req := PaymentRequest{
		Amount:      amount,
		Currency:    currency,
		Destination: c.paymentAddress,
		Memo:        memo,
	}

	data, _ := json.Marshal(req)
	_ = data
	resp, err := c.httpClient.Post("https://pay.example.com/api/v1/payment", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, err
	}
	return &paymentResp, nil
}