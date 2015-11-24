package amazon

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

const (
	SandboxURL    string = "http://localhost:8080/RVSSandbox"
	ProductionURL string = "https://appstore-sdk.amazon.com"
)

// Config is a configuration to initialize client
type Config struct {
	IsProduction bool
	Secret       string
	TimeOut      time.Duration
}

// The IAPResponse type has the response properties
type IAPResponse struct {
	ReceiptID       string `json:"receiptId"`
	ProductType     string `json:"productType"`
	ProductID       string `json:"productId"`
	PurchaseDate    int64  `json:"purchaseDate"`
	CancelDate      int64  `json:"cancelDate"`
	TestTransaction bool   `json:"testTransaction"`
}

type IAPResponseError struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

// IAPClient is an interface to call validation API in Amazon App Store
type IAPClient interface {
	Verify(string, string) (IAPResponse, error)
}

// Client implements IAPClient
type Client struct {
	URL     string
	Secret  string
	TimeOut time.Duration
}

// New creates a client object
func New(secret string) IAPClient {
	client := Client{
		URL:     SandboxURL,
		Secret:  secret,
		TimeOut: time.Second * 5,
	}
	if os.Getenv("IAP_ENVIRONMENT") == "production" {
		client.URL = ProductionURL
	}
	return client
}

// NewWithConfig creates a client with configuration
func NewWithConfig(config Config) Client {
	if config.TimeOut == 0 {
		config.TimeOut = time.Second * 5
	}

	client := Client{
		URL:     SandboxURL,
		Secret:  config.Secret,
		TimeOut: config.TimeOut,
	}
	if config.IsProduction {
		client.URL = ProductionURL
	}

	return client
}

// Verify sends receipts and gets validation result
func (c Client) Verify(userID string, receiptID string) (IAPResponse, error) {
	result := IAPResponse{}
	url := fmt.Sprintf("%v/version/1.0/verifyReceiptId/developer/%v/user/%v/receiptId/%v", c.URL, c.Secret, userID, receiptID)
	res, body, errs := gorequest.New().
		Get(url).
		Timeout(c.TimeOut).
		End()

	if errs != nil {
		return result, fmt.Errorf("%v", errs)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		responseError := IAPResponseError{}
		json.NewDecoder(strings.NewReader(body)).Decode(&responseError)
		return result, errors.New(responseError.Message)
	}

	err := json.NewDecoder(strings.NewReader(body)).Decode(&result)

	return result, err
}
