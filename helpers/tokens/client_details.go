package tokens

import (
	"fmt"
	"os"
	"strings"

	"github.com/sploders101/nikola-telemetry/helpers/config"
)

type ClientManager struct {
	config config.ConfigFile
}

func NewClientManager(config config.ConfigFile) ClientManager {
	return ClientManager {
		config: config,
	}
}

func (self ClientManager) GetClientId() (string, error) {
	clientIdRaw, err := os.ReadFile(self.config.TeslaClientIdFile)
	if err != nil {
		return "", fmt.Errorf("Error reading client ID: %w", err)
	}
	return strings.TrimSpace(string(clientIdRaw)), nil
}

func (self ClientManager) GetClientSecret() (string, error) {
	clientSecretRaw, err := os.ReadFile(self.config.TeslaClientSecretFile)
	if err != nil {
		return "", fmt.Errorf("Error reading client secret: %w", err)
	}
	return strings.TrimSpace(string(clientSecretRaw)), nil
}
