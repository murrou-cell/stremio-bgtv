package config

import (
	"encoding/base64"
	"os"

	"main.go/types"

	"gopkg.in/yaml.v2"
)

func LoadConfig() (types.Config, error) {
	var config types.Config
	data, err := os.ReadFile("config/config.yml")
	if err != nil {
		return config, err
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	if config.Base != "" {
		decodedBase, err := base64.StdEncoding.DecodeString(config.Base)
		if err != nil {
			return config, err
		}
		config.Base = string(decodedBase)
	}
	return config, nil
}
