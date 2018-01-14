package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/jtrotsky/poynt-pay/manager"
	"github.com/jtrotsky/poynt-pay/poyntcloud/auth"
	"github.com/jtrotsky/poynt-pay/poyntcloud/config"
)

// Message contains message body content of a CloudMessage post
type Message struct {
	BusinessID string `json:"businessId,omitempty"`
	// StoreID           string    `json:"storeId,omitempty"`
	MessageExpiryTime int64 `json:"ttl,omitempty"` // This is a time until the message expires.
	// Recipient         Recipient `json:"recipient,omitempty"`
	Data string `json:"data"`
}

// Recipient contains application information that is expected to receive the cloud message
type Recipient struct {
	ClassName   string `json:"className,omitempty"`
	PackageName string `json:"packageName,omitempty"`
}

// Payment is the payment information required for the payment fragment payload
type Payment struct {
	Action         string `json:"action"`
	IsDebit        bool   `json:"isDebit,omitempty"`
	PurchaseAmount int64  `json:"purchaseAmount"`
	TipAmount      int64  `json:"tipAmount"`
	CurrencyCode   string `json:"currency"`
	ReferenceID    string `json:"referenceId"`
	OrderID        string `json:"orderId"`
	CallBackURL    string `json:"callbackUrl"`
}

// SendCloudMessage sends a message to the POYNT cloud which passes that message
// on to an application running on the POYNT device.
func SendCloudMessage(config *config.Configuration, creds *auth.OAuthCreds,
	paymentAmount string) error {

	// First check if we expect an expired token.
	if creds.Expiry.IsZero() && creds.Expiry.Unix() < time.Now().Unix() {
		fmt.Println("Token Expired")
		return errors.New("Token expired.")
	}

	// Convert amount string to float64
	paymentAmountFloat, err := strconv.ParseFloat(paymentAmount, 64)
	if err != nil {
		fmt.Println("error converting payment amount string to number:", err)
	}

	// Create UUID for payment referenceID
	referenceID := poyntcloud.GenerateReferenceID()
	var paymentData = Payment{
		Action:  "sale",
		IsDebit: true,
		// Convert amounts to int as only Java long accepted. Last two digits are
		// assumed to be cents, hence * 100.
		PurchaseAmount: int64(paymentAmountFloat * 100),
		TipAmount:      0, // We don't tip in New Zealand.
		// TipAmount:      int64(paymentAmountFloat * 0.20),
		CurrencyCode: "NZD",
		ReferenceID:  referenceID,
		OrderID:      "test-order-from-joe",
		CallBackURL:  "https://736ed89f.ngrok.com/callback",
	}
	paymentDataJSON, err := json.Marshal(&paymentData)
	if err != nil {
		fmt.Println("error marshalling payment data:", err)
	}
	data := bytes.NewBuffer(paymentDataJSON)

	cloudMessage := &Message{
		BusinessID:        config.BusinessID,
		MessageExpiryTime: 30, // 30 second timeout.
		Data:              data.String(),
	}

	cloudMessageURL := config.PoyntAPIHostURL + "/cloudMessages"
	messagePayload, err := json.Marshal(&cloudMessage)
	if err != nil {
		fmt.Println("Error marshalling message JSON:", err)
	}

	fmt.Println("-----------------------------------------------------------")
	fmt.Printf("\nSending CloudMessage to POYNT\n")
	fmt.Printf("MESSAGE:\n %s", messagePayload)

	req, err := http.NewRequest("POST", cloudMessageURL, bytes.NewBuffer(messagePayload))
	req.Header.Set("Authorization", creds.TokenType+" "+creds.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	// Create UUID for requestID
	requestID := poyntcloud.GenerateReferenceID()
	req.Header.Set("Poynt-Request-Id", requestID)
	req.Header.Set("User-Agent", "Poynt-Pay")

	// NOTE: for debug
	fmt.Printf("\n\nREQUEST ID: %s", requestID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error performing HTTP request:", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	fmt.Printf("\n\nRESPONSE:\n%d %s\n", resp.StatusCode, body)

	if resp.StatusCode == 401 {
		var authResponse = auth.Response{}
		err := json.Unmarshal(body, authResponse)
		if err != nil {
			fmt.Println("Error unmarshalling response payload:", err)
		}
		if authResponse.Code == "INVALID_ACCESS_TOKEN" {
			return errors.New("Invalid access token. Probably expired.")
		}
		url := auth.BuildOAuthURL(config)
		fmt.Println("Please visit and authorize application at:", url)
		return err
	}
	return err
}
