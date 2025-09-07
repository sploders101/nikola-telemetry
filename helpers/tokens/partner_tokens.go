package tokens

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sploders101/nikola-telemetry/helpers/config"
)

type TokenManager struct {
	config          config.ConfigFile
	clientManager   ClientManager
	mutex           sync.Mutex
	activeToken     string
	tokenExpiration time.Time
}

func NewTokenManager(config config.ConfigFile, clientManager ClientManager) *TokenManager {
	return &TokenManager{
		config: config,
		clientManager: clientManager,
	}
}

func (self *TokenManager) GetPartnerToken() (string, error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.tokenExpiration.After(time.Now()) || self.activeToken != "" {
		return self.activeToken, nil
	}

	clientId, err := self.clientManager.GetClientId()
	if err != nil {
		return "", err
	}

	clientSecret, err := self.clientManager.GetClientSecret()
	if err != nil {
		return "", err
	}

	response, err := http.PostForm(*self.config.TeslaRegistrationUrl, url.Values(map[string][]string{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"audience":      {self.config.TeslaBaseUrl},
		"scope":         {"openid user_data vehicle_device_data vehicle_cmds vehicle_charging_cmds"},
	}))
	if err != nil {
		return "", fmt.Errorf("Error sending http request: %w", err)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading http response body: %w", err)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("Received %v status code from token API", response.StatusCode)
	}

	var token struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(bodyBytes, &token); err != nil {
		return "", fmt.Errorf("Error unmarshaling json: %w", err)
	}

	if token.TokenType != "Bearer" {
		return "", fmt.Errorf("Unrecognized token type: %v", token.TokenType)
	}

	self.activeToken = token.AccessToken
	tokenExpiration := time.Duration(token.ExpiresIn) * time.Second
	self.tokenExpiration = time.Now().Add(tokenExpiration - (time.Duration(1) * time.Minute))

	return token.AccessToken, nil
}
