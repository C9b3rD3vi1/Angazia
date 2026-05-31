package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
)

type IntaSendClient struct {
	apiKey     string
	apiSecret  string
	publishableKey string
	baseURL    string
	httpClient *http.Client
}

type IntaSendChargeRequest struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phone_number,omitempty"`
	Reference   string  `json:"reference"`
	Narrative   string  `json:"narrative"`
	WebhookURL  string  `json:"webhook_url"`
	RedirectURL string  `json:"redirect_url"`
}

type IntaSendChargeResponse struct {
	Status       string `json:"status"`
	InvoiceID    string `json:"invoice_id"`
	RedirectURL  string `json:"redirect_url"`
	TransactionID string `json:"transaction_id"`
	Message      string `json:"message"`
}

type IntaSendPaymentStatusResponse struct {
	Status       string  `json:"status"`
	TransactionID string `json:"transaction_id"`
	Reference    string  `json:"reference"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	PaymentMethod string `json:"payment_method"`
	CreatedAt    string  `json:"created_at"`
}

type IntaSendWebhookData struct {
	Event         string                 `json:"event"`
	TransactionID string                 `json:"transaction_id"`
	Reference     string                 `json:"reference"`
	Status        string                 `json:"status"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
}

func NewIntaSendClient(cfg *config.Config) *IntaSendClient {
	baseURL := "https://api.intasend.com/v1"
	if cfg.Environment == "development" {
		baseURL = "https://sandbox.intasend.com/api/v1"
	}
	
	return &IntaSendClient{
		apiKey:         cfg.IntaSendAPIKey,
		apiSecret:      cfg.IntaSendAPISecret,
		publishableKey: cfg.IntaSendPublishableKey,
		baseURL:        baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *IntaSendClient) CreateCharge(request *IntaSendChargeRequest) (*IntaSendChargeResponse, error) {
	url := c.baseURL + "/payments/mpesa/stk_push/"
	
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("intaSend API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var response IntaSendChargeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &response, nil
}

func (c *IntaSendClient) GetPaymentStatus(transactionID string) (*IntaSendPaymentStatusResponse, error) {
	url := fmt.Sprintf("%s/payments/%s/status/", c.baseURL, transactionID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("intaSend API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var response IntaSendPaymentStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &response, nil
}

func (c *IntaSendClient) VerifyWebhookSignature(payload []byte, signature string) bool {
	hash := hmac.New(sha256.New, []byte(c.apiSecret))
	hash.Write(payload)
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (c *IntaSendClient) RefundPayment(transactionID string, amount float64) error {
	url := c.baseURL + "/payments/refund/"
	
	refundRequest := map[string]interface{}{
		"transaction_id": transactionID,
		"amount":         amount,
	}
	
	jsonBody, err := json.Marshal(refundRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("intaSend API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}