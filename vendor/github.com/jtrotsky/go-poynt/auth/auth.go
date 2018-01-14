package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jtrotsky/go-poynt/poyntcloud"
	"github.com/jtrotsky/go-poynt/poyntcloud/config"
)

// GetAuth gets OAuth credentials using configuration
func GetAuth(config *config.Configuration) (*OAuthCreds, error) {
	creds, err := getAccessToken(config)
	if err != nil {
		fmt.Println("Error getting access token:", err)
	}
	return creds, err
}

// Response is the HTTP response from the POYNT cloud server.
type Response struct {
	Code             string `json:"code,omitempty"`
	Status           int    `json:"httpStatus,omitempty"`
	Message          string `json:"message,omitempty"`
	DeveloperMessage string `json:"developerMessage,omitempty"`
	RequestID        string `json:"requestId,omitempty"`
}

// OAuthCreds contains authentication data returned from JWT auth request
type OAuthCreds struct {
	AccessToken  string `json:"accessToken,omitempty"`
	TokenType    string `json:"tokenType,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int64  `json:"expiresIn,omitempty"`
	Scope        string `json:"scope,omitempty"`
	// My own parameter to store expiry time.
	Expiry time.Time
}

// getAccessToken retrieves access token from POYNT services.
func getAccessToken(config *config.Configuration) (*OAuthCreds, error) {
	fmt.Println("Generating JWT Token")
	tokenString, err := genJWTToken(config)
	if err != nil {
		fmt.Println("Error generating JWT token:", err)
	}

	// Add request parameters.
	params := url.Values{}
	params.Add("grantType", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	params.Add("assertion", tokenString)

	body, statusCode, err := authRequest(params, config)
	if err != nil {
		fmt.Println("Error performing authentication request:", err)
	}

	// TODO: ordering
	if statusCode != 200 {
		fmt.Printf("Bad HTTP response: %s %d", body, statusCode)
	} else {
		fmt.Println("Token received with status:", statusCode)
		fmt.Println("Ready to send messages to application")
	}

	Creds := OAuthCreds{}
	if json.Unmarshal(body, &Creds); err != nil {
		fmt.Println("error unmarshalling response into OAuthCreds:", err)
	}

	// Set expiry to the current time plus the given expiresIn time.
	Creds.Expiry = time.Now().Add(time.Duration(Creds.ExpiresIn * 1000))

	return &Creds, err
}

// RefreshAccessToken refreshes the OAuth access token.
func RefreshAccessToken(config *config.Configuration, creds *OAuthCreds) (*OAuthCreds, error) {
	// Add request parameters.
	params := url.Values{}
	params.Add("grantType", "REFRESH_TOKEN")
	params.Add("refreshToken", creds.RefreshToken)

	body, statusCode, err := authRequest(params, config)
	if err != nil {
		fmt.Println("Error performing authentication request:", err)
	}

	if statusCode != 200 {
		fmt.Printf("Bad HTTP response: %s %d", body, statusCode)
	} else {
		fmt.Println("Refresh token received with status:", statusCode)
		fmt.Println("Ready to send messages to application")
	}

	Creds := OAuthCreds{}
	if json.Unmarshal(body, &Creds); err != nil {
		fmt.Println("error unmarshalling response into OAuthCreds:", err)
	}

	// Set expiry to the current time plus the given expiresIn time.
	Creds.Expiry = time.Now().Add(time.Duration(Creds.ExpiresIn * 1000))

	return &Creds, err
}

func authRequest(params url.Values, config *config.Configuration) ([]byte, int, error) {
	tokenURL := "https://services.poynt.net" + "/token"
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	req.Header.Set("api-version", strconv.FormatFloat(config.PoyntAPIVersion, 'f', 1, 64))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Create UUID for requestID
	requestID := poyntcloud.GenerateReferenceID()
	req.Header.Set("Poynt-Request-Id", requestID)
	req.Header.Set("User-Agent", "go-poynt")

	client := &http.Client{}
	fmt.Println("Requesting access token from POYNT")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error performing HTTP request:", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	return body, resp.StatusCode, err
}

func genJWTToken(config *config.Configuration) (string, error) {
	// Create JWT token
	token := jwt.New(jwt.SigningMethodRS256)

	// Set some default claims
	token.Claims["iss"] = config.ApplicationID   // Issuer
	token.Claims["sub"] = config.ApplicationID   // Subject
	token.Claims["aud"] = config.PoyntAPIHostURL // Audience
	token.Claims["exp"] = (fiveMinutesFromNow()) // Expiry time
	token.Claims["iat"] = time.Now().Unix()      // Time JWT issued
	// Create UUID for request reference
	referenceID := poyntcloud.GenerateReferenceID()
	token.Claims["jti"] = referenceID // Unique ID.

	// Read private key from config file
	mySigningKey, err := ioutil.ReadFile(config.PrivateKeyFile)
	if err != nil {
		fmt.Println("error reading private key from file:", err)
	}
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Println("error signing token with key:", err)
	}
	return tokenString, err
}

// To calculate token expiry times
func fiveMinutesFromNow() int64 {
	timeNow := time.Now()

	// Minutes til expiry
	minutes := 5
	fiveMinutes := time.Duration(minutes) * time.Minute

	timeInFiveMinutes := timeNow.Add(fiveMinutes)
	return timeInFiveMinutes.Unix()
}

// BuildOAuthURL creates the OAuth URL for application approval
func BuildOAuthURL(config *config.Configuration) string {
	var address string
	baseURL := "https://poynt.net/applications/"

	query := url.Values{}
	query.Add("callback", fmt.Sprintf("%s", "https://736ed89f.ngrok.com/callback"))
	query.Add("applicationId", fmt.Sprintf("%s", config.ApplicationID))
	// query.Add("context", fmt.Sprintf("%s", "go-poynt"))

	address += fmt.Sprintf("%sauthorize?%s", baseURL, query.Encode())
	return address
}
