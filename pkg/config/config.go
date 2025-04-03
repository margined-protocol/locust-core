package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/margined-protocol/locust-core/pkg/types"
	"github.com/mitchellh/mapstructure"
)

func LoadConfig(configPath string) (*types.Config, error) {
	// Ensure the file exists and is readable
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at path: %s", configPath)
	}

	// Open and read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the TOML data
	var cfg types.Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Convert retry delay to milliseconds
	cfg.TxRetryDelay *= time.Millisecond

	return &cfg, nil
}

func DecodeConfig(input interface{}, output interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook:       types.SdkIntDecodeHook,
		Result:           output,
		WeaklyTypedInput: true, // Allows string-to-number conversions
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("failed to create decoder: %w", err)
	}

	return decoder.Decode(input)
}
