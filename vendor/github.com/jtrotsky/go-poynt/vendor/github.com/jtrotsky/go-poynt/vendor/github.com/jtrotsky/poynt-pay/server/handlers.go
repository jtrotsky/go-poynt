package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/jtrotsky/poynt-pay/poyntcloud"
	"github.com/jtrotsky/poynt-pay/poyntcloud/actions/message"
	"github.com/jtrotsky/poynt-pay/poyntcloud/auth"
	"github.com/jtrotsky/poynt-pay/poyntcloud/config"
)

// TODO: Separate callback for OAuth callback as opposed to cloudMessage callback
// The info returned from POYNT to specified callback URI.
type callbackResult struct {
	ReferenceID string `json:"referenceId"`
	Status      string `json:"status"`
}

var (
	// String is the ID, and chan is the channel
	callbacks = map[string]chan callbackResult{}
	// Mutex handles locking maps, so can only be accessed by one CPU
	callbackMutex = sync.Mutex{}
)

// Manager stores credentials and configuration for a given store/user.
type Manager struct {
	Creds  *auth.OAuthCreds
	Config *config.Configuration
}

// NewManager creates a manager that contains credentials and configuration for
// a user.
func NewManager(Auth *auth.OAuthCreds, Config *config.Configuration) *Manager {
	return &Manager{Auth, Config}
}

// Gateway is the basic landing page.
func (manager *Manager) Gateway(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "gateway.html", nil)
	w.WriteHeader(http.StatusOK)
}

// Pay sends a payment to POYNT and waits for a response.
func (manager *Manager) Pay(w http.ResponseWriter, r *http.Request) {
	// TODO: where to write header?
	w.WriteHeader(http.StatusOK)

	var err error
	// Default amount to send.
	var amountParam = "00.00"
	r.ParseForm()
	// Capture sale "amount" and "origin" passed as query parameters from Vend.
	for key, param := range r.Form {
		// TODO:
		// // Log origin. Need to eventually compare this to the domain prefix we expect.
		if key == "origin" && param != nil {
			log.Println("Origin:", param[0])
		}
		if key == "amount" && param != nil {
			amountParam = param[0]
		}
	}

	// TODO: For debug
	fmt.Println("Amount received:", amountParam)

	// Convert amount string to float64 and check it's > 0.01.
	paymentAmount, err := strconv.ParseFloat(amountParam, 64)
	if err != nil {
		fmt.Println("Error converting payment amount string to number:", err)
	}
	if paymentAmount < 0.01 {
		// TODO: Fail elegantly.
		// Return transaction failed, with reason why?
		panic(err)
	}

	// Make call to poynt terminal.
	// Generate UUID to identify transaction.
	referenceID := poyntcloud.GenerateReferenceID()
	// Channel expects a result
	ch := make(chan callbackResult)

	// Lock prevents reading from maps at same time.
	callbackMutex.Lock()
	// Create channel with our unique ID
	callbacks[referenceID] = ch
	callbackMutex.Unlock()

	// Send amount to POYNT terminal.
	// Check if cloud message sends successfully, if it doesn't then retry (most common
	// cause being access token needing refresh).
	if message.SendCloudMessage(manager.Config, manager.Creds, paymentAmount, referenceID); err != nil {
		// TODO for debug
		fmt.Println("Refreshing access token")
		auth, err := auth.RefreshAccessToken(manager.Config, manager.Creds)
		if err != nil {
			fmt.Println("Error refreshing access token:", err)
		}
		// TODO for debug
		fmt.Println("Sending cloud message again")
		if message.SendCloudMessage(manager.Config, auth, paymentAmount, referenceID); err != nil {
			log.Fatalf("Failed to send cloud message twice: %v", err)
		}
	}

	// Wait until the channel gets a result from callback.
	res := <-ch
	// Turn response struct into JSON.
	resJSON, err := json.MarshalIndent(res, "", "\t")
	// Return to the AJAX call from the frontend.
	w.Write(resJSON)

	// Finally, remove the channel for memory's sake.
	callbackMutex.Lock()
	delete(callbacks, referenceID)
	callbackMutex.Unlock()
}

// Callback is a URL that listens for the POYNT terminals response messages.
func (manager *Manager) Callback(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	messageResponse := callbackResult{}
	err := decoder.Decode(&messageResponse)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nUSER ACTION: \n%s\n", messageResponse)

	res := callbackResult{
		// Status, reference.
		ReferenceID: messageResponse.ReferenceID,
		Status:      messageResponse.Status,
	}

	callbackMutex.Lock()
	// Check callback[id] exists
	ch, ok := callbacks[res.ReferenceID]
	callbackMutex.Unlock()
	if !ok {
		// Log and wonder wtf happend
		// why did we never send that transaction
		log.Printf("Error, couldn't find ID for chan: %v", err)
	}
	// receive result on that channel
	ch <- res
}
