package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Configuration is for API and application configuration
type Configuration struct {
	PackageName        string  `json:"package_name,omitempty"`          // com.vendhq.poyntlisten
	ClassName          string  `json:"class_name,omitempty"`            // com.vendhq.poyntlisten.MainReceiverClass
	PoyntAPIHostURL    string  `json:"poynt_api_host_url,omitempty"`    // https://services.poynt.net
	PoyntAPIVersion    float64 `json:"poynt_api_version,omitempty"`     // 1.2
	PoyntAuthHostURL   string  `json:"poynt_auth_host_url,omitempty"`   // https://poynt.net
	BusinessID         string  `json:"business_id,omitempty"`           // c58ceb6f-3ecb-4000-84cf-f981f34ce482 Honest Mulch
	StoreID            string  `json:"store_id,omitempty"`              // fa937f9f-4493-4941-bded-7c2db42e8c9a Honest Mulch 458
	ApplicationID      string  `json:"application_id,omitempty"`        // urn:aid:67dae7d1-a503-443d-a000-6da70bc98743 Poynt Pay
	DeviceID           string  `json:"device_id,omitempty"`             // l4zo
	PrivateKeyFile     string  `json:"private_key_file,omitempty"`      // keys/poynt_pay_key
	PublicKeyFile      string  `json:"public_key_file,omitempty"`       // keys/poynt_pay_key.pub
	PoyntPublicKeyFile string  `json:"poynt_public_key_file,omitempty"` // keys/services.poynt.net.pub
}

// GetConfig creates a Configuration object from config JSON
func GetConfig() (*Configuration, error) {
	// Load configuration from ./config/conf.json
	config, err := loadConfig()
	return config, err
}

// loadConfig reads and stores config from ./config/conf.json
func loadConfig() (*Configuration, error) {
	// Read config from file
	file, err := ioutil.ReadFile("./poyntcloud/config/conf.json")
	if err != nil {
		fmt.Println("error opening and reading config file:", err)
	}
	config := Configuration{}
	// Unmarshal config into Configuration struct
	json.Unmarshal(file, &config)
	return &config, err
}
