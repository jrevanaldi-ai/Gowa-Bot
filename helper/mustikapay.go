package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)


const (
	MustikaPayBaseURL = "https://mustikapayment.com"
	MustikaPayCreateEndpoint = "/api/createpay"
	MustikaPayCheckEndpoint  = "/api/cekpay"
	DefaultMDR               = 0.7
)


type MustikaPayClient struct {
	APIKey  string
	Logger  *Logger
	Client  *http.Client
}


type CreatePaymentRequest struct {
	Amount       int    `form:"amount"`
	ProductName  string `form:"product_name"`
	CustomerName string `form:"customer_name"`
	RedirectURL  string `form:"redirect_url"`
}


type CreatePaymentResponse struct {
	Status       json.RawMessage `json:"status"`
	StatusStr    string          `json:"-"`
	Message      string          `json:"message"`
	RefNo        string          `json:"ref_no"`
	QRString     string          `json:"qr_string"`
	QRContent   string          `json:"qr_content"`
	QRImageURL   string          `json:"qr_image"`
	QRURL        string          `json:"qr_url"`
	PaymentLink  string          `json:"payment_link"`
	Amount       json.Number     `json:"amount"`
	ProductName  string          `json:"product_name"`
	ExpiresAt    string          `json:"expires_at"`
	Type         string          `json:"type"`
}


func (r *CreatePaymentResponse) UnmarshalJSON(data []byte) error {

	type Alias CreatePaymentResponse
	aux := &struct {
		Status json.RawMessage `json:"status"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.Status = aux.Status

	if r.Status != nil {
		s := strings.Trim(string(r.Status), "\"")
		r.StatusStr = strings.ToLower(s)
	}

	return nil
}


func (r *CreatePaymentResponse) GetQRString() string {
	
	if r.QRContent != "" {
		return r.QRContent
	}
	if r.QRString != "" {
		return r.QRString
	}
	return ""
}


func (r *CreatePaymentResponse) GetQRImageURL() string {
	
	if r.QRURL != "" {
		return r.QRURL
	}
	if r.QRImageURL != "" {
		return r.QRImageURL
	}
	return ""
}


func (r *CreatePaymentResponse) IsSuccess() bool {

	if r.StatusStr != "" {
		return r.StatusStr == "success" || r.StatusStr == "true" || r.StatusStr == "1" || r.StatusStr == "ok"
	}

	statusStr := string(r.Status)

	if statusStr == "" || statusStr == "null" {
		return false
	}

	s := strings.ToLower(strings.Trim(statusStr, "\""))
	return s == "success" || s == "true" || s == "1" || s == "ok"
}


func (r *CreatePaymentResponse) GetAmountInt() int {
	if r.Amount == "" {
		return 0
	}
	
	amount, err := r.Amount.Int64()
	if err != nil {
		return 0
	}
	return int(amount)
}


type CheckPaymentResponse struct {
	RefNo  string `json:"ref_no"`
	Status string `json:"status"`
	Type   string `json:"type"`
	Amount int    `json:"amount"`
	Issuer string `json:"issuer"`
	Payor  string `json:"payor"`
}


type PaymentStatus string

const (
	StatusSuccess PaymentStatus = "success"
	StatusPending PaymentStatus = "pending"
	StatusExpired PaymentStatus = "expired"
	StatusFailed  PaymentStatus = "failed"
)


func NewMustikaPayClient(apiKey string, logger *Logger) *MustikaPayClient {
	return &MustikaPayClient{
		APIKey: apiKey,
		Logger: logger,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}


func (c *MustikaPayClient) CreateQRIS(req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	if req.ProductName == "" {
		req.ProductName = "Donasi Gowa-Bot"
	}
	if req.CustomerName == "" {
		req.CustomerName = "Donatur"
	}


	formData := url.Values{}
	formData.Set("amount", fmt.Sprintf("%d", req.Amount))
	formData.Set("product_name", req.ProductName)
	formData.Set("customer_name", req.CustomerName)
	if req.RedirectURL != "" {
		formData.Set("redirect_url", req.RedirectURL)
	}


	apiURL := MustikaPayBaseURL + MustikaPayCreateEndpoint


	httpReq, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}


	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("X-Api-Key", c.APIKey)
	httpReq.Header.Set("User-Agent", "Gowa-Bot/1.0")


	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()


	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}


	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}


	var apiResp CreatePaymentResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}


	if !apiResp.IsSuccess() {
		return nil, fmt.Errorf("API error: %s", apiResp.Message)
	}

	amount := apiResp.GetAmountInt()
	c.Logger.Info("QRIS created: %s (Amount: %d)", apiResp.RefNo, amount)

	return &apiResp, nil
}


func (c *MustikaPayClient) CheckPaymentStatus(refNo string) (*CheckPaymentResponse, error) {
	if refNo == "" {
		return nil, fmt.Errorf("ref_no cannot be empty")
	}


	apiURL := fmt.Sprintf("%s%s?ref_no=%s", MustikaPayBaseURL, MustikaPayCheckEndpoint, url.QueryEscape(refNo))


	httpReq, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}


	httpReq.Header.Set("X-Api-Key", c.APIKey)
	httpReq.Header.Set("User-Agent", "Gowa-Bot/1.0")


	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()


	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}


	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}


	var statusResp CheckPaymentResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}


	if statusResp.RefNo == "" {
		return nil, fmt.Errorf("invalid response: ref_no not found")
	}

	c.Logger.Debug("Payment status checked: %s - %s", refNo, statusResp.Status)

	return &statusResp, nil
}


func (c *MustikaPayClient) WaitForPayment(refNo string, interval time.Duration, maxAttempts int) (*CheckPaymentResponse, error) {
	if interval == 0 {
		interval = 5 * time.Second
	}
	if maxAttempts == 0 {
		maxAttempts = 24
	}

	c.Logger.Info("Start polling payment %s (interval: %v, max attempts: %d)", refNo, interval, maxAttempts)

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(interval)

		status, err := c.CheckPaymentStatus(refNo)
		if err != nil {
			c.Logger.Warning("Failed to check payment status (attempt %d/%d): %v", i+1, maxAttempts, err)
			continue
		}


		if PaymentStatus(status.Status) == StatusSuccess {
			c.Logger.Info("Payment successful: %s", refNo)
			return status, nil
		}


		if PaymentStatus(status.Status) == StatusExpired || PaymentStatus(status.Status) == StatusFailed {
			c.Logger.Info("Payment expired/failed: %s", refNo)
			return status, nil
		}


		c.Logger.Debug("Payment still pending: %s (attempt %d/%d)", refNo, i+1, maxAttempts)
	}

	return nil, fmt.Errorf("payment polling timeout after %d attempts", maxAttempts)
}


func (s PaymentStatus) IsValid() bool {
	switch s {
	case StatusSuccess, StatusPending, StatusExpired, StatusFailed:
		return true
	default:
		return false
	}
}


func (s PaymentStatus) IsSuccess() bool {
	return s == StatusSuccess
}


func (s PaymentStatus) IsPending() bool {
	return s == StatusPending
}


func (s PaymentStatus) IsExpired() bool {
	return s == StatusExpired
}


func FormatAmount(amount int) string {

	if amount >= 1000000 {
		return fmt.Sprintf("%d.%03d.%03d", amount/1000000, (amount/1000)%1000, amount%1000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%d.%03d", amount/1000, amount%1000)
	}
	return fmt.Sprintf("%d", amount)
}
