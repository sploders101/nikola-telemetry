package config

import (
	"encoding/json"
	"os"
	"path"
)

type ConfigFile struct {
	// Contains the base path for all relative file paths in the config.
	// If unset, this will default to the parent directory of the config.
	BasePath *string

	// The listening address for data that should be hosted publicly
	// (namely the `.well-known` folder). If this is null, the private
	// address will be used for all hosting. Please make sure to whitelist
	// the `.well-known` folder in your reverse proxy.
	PublicAddress *string

	// The listening address for the APIs that this proxy provides
	PrivateAddress string

	// The domain that we should use to register our application
	Domain string

	// The URL for registering to a partner API. This is different for Chinese users.
	// Default is https://fleet-auth.prd.vn.cloud.tesla.com/oauth2/v3/token
	TeslaRegistrationUrl *string

	// The base URL for your region to use when calling the Tesla API
	TeslaBaseUrl string

	// Path to file containing the private key for command signing
	TeslaPrivkeyFile string

	// Path to the file containing public key for command signing
	TeslaPubkeyFile string

	// Path to the file containing the client ID for your API application
	TeslaClientIdFile string

	// Path to the file containing the client secret for your API application
	TeslaClientSecretFile string
}

func LoadConfig() (ConfigFile, error) {
	filePath := "./config/config.json"
	if env := os.Getenv("CONFIG_PATH"); env != "" {
		filePath = env
	}

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return ConfigFile{}, err
	}

	var config ConfigFile
	if err := json.Unmarshal(contents, &config); err != nil {
		return ConfigFile{}, err
	}

	// Add BasePath if it is unspecified
	if config.BasePath == nil {
		basePath := path.Dir(filePath)
		config.BasePath = &basePath
	}

	if !path.IsAbs(config.TeslaPrivkeyFile) {
		config.TeslaPrivkeyFile = path.Join(*config.BasePath, config.TeslaPrivkeyFile)
	}
	if !path.IsAbs(config.TeslaPubkeyFile) {
		config.TeslaPubkeyFile = path.Join(*config.BasePath, config.TeslaPubkeyFile)
	}
	if !path.IsAbs(config.TeslaClientIdFile) {
		config.TeslaClientIdFile = path.Join(*config.BasePath, config.TeslaClientIdFile)
	}
	if !path.IsAbs(config.TeslaClientSecretFile) {
		config.TeslaClientSecretFile = path.Join(*config.BasePath, config.TeslaClientSecretFile)
	}

	// Add defaults
	if config.TeslaRegistrationUrl == nil {
		defaultUrl := "https://fleet-auth.prd.vn.cloud.tesla.com/oauth2/v3/token"
		config.TeslaRegistrationUrl = &defaultUrl
	}

	return config, nil
}

