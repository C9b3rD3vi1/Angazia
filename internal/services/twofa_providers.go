package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
)

// SMSProvider interface for SMS delivery
type SMSProvider interface {
	Send(phoneNumber, message string) error
	SendWithTemplate(phoneNumber, templateID string, data map[string]string) error
	GetProviderName() string
}

// ========== AFRICA'S TALKING SMS PROVIDER ==========

type AfricaTalkingSMSProvider struct {
	apiKey     string
	username   string
	from       string
	baseURL    string
	httpClient *http.Client
}

func NewAfricaTalkingSMSProvider(cfg *config.Config) *AfricaTalkingSMSProvider {
	baseURL := "https://api.africastalking.com/version1"
	if cfg.Environment == "development" {
		baseURL = "https://api.sandbox.africastalking.com/version1"
	}
	
	return &AfricaTalkingSMSProvider{
		apiKey:   cfg.AfricaTalkingAPIKey,
		username: cfg.AfricaTalkingUsername,
		from:     cfg.AfricaTalkingFrom,
		baseURL:  baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *AfricaTalkingSMSProvider) Send(phoneNumber, message string) error {
	url := p.baseURL + "/messaging"
	
	// Format phone number to international format if needed
	formattedNumber := p.formatPhoneNumber(phoneNumber)
	
	data := map[string]string{
		"username": p.username,
		"to":       formattedNumber,
		"message":  message,
		"from":     p.from,
	}
	
	jsonData, _ := json.Marshal(data)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", p.apiKey)
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("SMS API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	if smsData, ok := result["SMSMessageData"].(map[string]interface{}); ok {
		if recipients, ok := smsData["Recipients"].([]interface{}); ok && len(recipients) > 0 {
			if recipient, ok := recipients[0].(map[string]interface{}); ok {
				if status, ok := recipient["status"].(string); ok && status != "Success" {
					return fmt.Errorf("SMS delivery failed: %v", recipient["status"])
				}
			}
		}
	}
	
	return nil
}

func (p *AfricaTalkingSMSProvider) SendWithTemplate(phoneNumber, templateID string, data map[string]string) error {
	// For Africa's Talking, you'd need to implement template rendering
	// This is a simplified version
	message := fmt.Sprintf("Your verification code is: %s", data["code"])
	return p.Send(phoneNumber, message)
}

func (p *AfricaTalkingSMSProvider) GetProviderName() string {
	return "africastalking"
}

func (p *AfricaTalkingSMSProvider) formatPhoneNumber(phone string) string {
	// Remove any non-digit characters
	cleaned := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			cleaned += string(c)
		}
	}
	
	// Kenyan numbers: start with 07 or 01, need to convert to 2547 or 2541
	if len(cleaned) == 10 && (cleaned[:2] == "07" || cleaned[:2] == "01") {
		return "254" + cleaned[1:]
	}
	
	// Already has 254 prefix
	if len(cleaned) == 12 && cleaned[:3] == "254" {
		return cleaned
	}
	
	// Has +254 prefix
	if len(cleaned) == 13 && cleaned[:4] == "254" {
		return cleaned
	}
	
	return cleaned
}

// ========== TWILIO SMS PROVIDER ==========

type TwilioSMSProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	baseURL    string
	httpClient *http.Client
}

func NewTwilioSMSProvider(cfg *config.Config) *TwilioSMSProvider {
	return &TwilioSMSProvider{
		accountSID: cfg.TwilioAccountSID,
		authToken:  cfg.TwilioAuthToken,
		fromNumber: cfg.TwilioFromNumber,
		baseURL:    "https://api.twilio.com/2010-04-01",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *TwilioSMSProvider) Send(phoneNumber, message string) error {
	url := fmt.Sprintf("%s/Accounts/%s/Messages.json", p.baseURL, p.accountSID)
	
	data := fmt.Sprintf("To=%s&From=%s&Body=%s", phoneNumber, p.fromNumber, message)
	
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.SetBasicAuth(p.accountSID, p.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Twilio API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (p *TwilioSMSProvider) SendWithTemplate(phoneNumber, templateID string, data map[string]string) error {
	message := fmt.Sprintf("Your verification code is: %s", data["code"])
	return p.Send(phoneNumber, message)
}

func (p *TwilioSMSProvider) GetProviderName() string {
	return "twilio"
}

// ========== VONAGE (NEXMO) SMS PROVIDER ==========

type VonageSMSProvider struct {
	apiKey     string
	apiSecret  string
	from       string
	baseURL    string
	httpClient *http.Client
}

func NewVonageSMSProvider(cfg *config.Config) *VonageSMSProvider {
	return &VonageSMSProvider{
		apiKey:    cfg.VonageAPIKey,
		apiSecret: cfg.VonageAPISecret,
		from:      cfg.VonageFrom,
		baseURL:   "https://rest.nexmo.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *VonageSMSProvider) Send(phoneNumber, message string) error {
	url := p.baseURL + "/sms/json"
	
	data := map[string]string{
		"api_key":    p.apiKey,
		"api_secret": p.apiSecret,
		"to":         phoneNumber,
		"from":       p.from,
		"text":       message,
	}
	
	formData := ""
	for k, v := range data {
		formData += fmt.Sprintf("%s=%s&", k, v)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(formData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	if messages, ok := result["messages"].([]interface{}); ok && len(messages) > 0 {
		if msg, ok := messages[0].(map[string]interface{}); ok {
			if status, ok := msg["status"].(string); ok && status != "0" {
				return fmt.Errorf("SMS failed: %v", msg["error-text"])
			}
		}
	}
	
	return nil
}

func (p *VonageSMSProvider) SendWithTemplate(phoneNumber, templateID string, data map[string]string) error {
	message := fmt.Sprintf("Your verification code is: %s", data["code"])
	return p.Send(phoneNumber, message)
}

func (p *VonageSMSProvider) GetProviderName() string {
	return "vonage"
}

// ========== MOCK SMS PROVIDER (FOR DEVELOPMENT) ==========

type MockSMSProvider struct{}

func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{}
}

func (p *MockSMSProvider) Send(phoneNumber, message string) error {
	fmt.Printf("[MOCK SMS] To: %s, Message: %s\n", phoneNumber, message)
	return nil
}

func (p *MockSMSProvider) SendWithTemplate(phoneNumber, templateID string, data map[string]string) error {
	message := fmt.Sprintf("Your verification code is: %s", data["code"])
	return p.Send(phoneNumber, message)
}

func (p *MockSMSProvider) GetProviderName() string {
	return "mock"
}

// ========== SMS PROVIDER FACTORY ==========

type SMSProviderFactory struct {
	cfg *config.Config
}

func NewSMSProviderFactory(cfg *config.Config) *SMSProviderFactory {
	return &SMSProviderFactory{cfg: cfg}
}

func (f *SMSProviderFactory) GetProvider() SMSProvider {
	provider := f.cfg.SMSProvider
	if provider == "" {
		provider = "mock"
	}
	
	switch provider {
	case "africastalking":
		return NewAfricaTalkingSMSProvider(f.cfg)
	case "twilio":
		return NewTwilioSMSProvider(f.cfg)
	case "vonage":
		return NewVonageSMSProvider(f.cfg)
	default:
		return NewMockSMSProvider()
	}
}