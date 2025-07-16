package types

import (
	"fmt"
	"reflect"
	"time"

	"github.com/BurntSushi/toml"

	sdkmath "cosmossdk.io/math"
)

type SigningKey struct {
	AppName string `toml:"app_name"`
	Backend string `toml:"backend"`
	RootDir string `toml:"root_dir"`
}

type Chain struct {
	Prefix        string               `toml:"prefix" mapstructure:"prefix"`
	ChainID       string               `toml:"chain_id" mapstructure:"chain_id"`
	Fees          *string              `toml:"fees" mapstructure:"fees"`
	Gas           *string              `toml:"gas" mapstructure:"gas"`
	GasAdjustment *float64             `toml:"gas_adjustment" mapstructure:"gas_adjustment"`
	GasPrices     *string              `toml:"gas_prices" mapstructure:"gas_prices"`
	GasDenom      string               `toml:"gas_denom" mapstructure:"gas_denom"`
	GRPCEndpoints []GRPCEndpointConfig `toml:"grpc_endpoints" mapstructure:"grpc_endpoints"`
	RPCEndpoints  []RPCEndpointConfig  `toml:"rpc_endpoints" mapstructure:"rpc_endpoints"`
}

type Config struct {
	Chain         Chain            `toml:"chain"`
	Key           SigningKey       `toml:"key"`
	Memo          string           `toml:"memo"`
	WebsocketPath string           `toml:"websocket_path"`
	SignerAccount string           `toml:"signer_account"`
	Strategies    []StrategyConfig `toml:"strategy"`
	TxRetryCount  int              `toml:"tx_retry_count"`
	TxRetryDelay  time.Duration    `toml:"tx_retry_delay_ms"`
	DryRun        bool             `toml:"dry_run"`
	PropTrade     bool             `toml:"prop_trade"`
}

type GRPCEndpointConfig struct {
	Address  string `toml:"grpc_server_address" mapstructure:"grpc_server_address"`
	UseTLS   bool   `toml:"grpc_tls" mapstructure:"grpc_tls"`
	APIToken string `toml:"grpc_api_token" mapstructure:"grpc_api_token"`
}

type RPCEndpointConfig struct {
	Address  string `toml:"rpc_server_address" mapstructure:"rpc_server_address"`
	APIToken string `toml:"rpc_api_token" mapstructure:"rpc_api_token"`
}

type LocustVault struct {
	Fund     string `toml:"fund" mapstructure:"fund"`
	Strategy string `toml:"strategy" mapstructure:"strategy"`
}

type StrategyConfig struct {
	Name   string
	Config interface{}
}

// SdkInt is a wrapper around sdkmath.Int to handle TOML unmarshalling
type SdkInt struct {
	Value sdkmath.Int
}

// UnmarshalTOML implements TOML unmarshalling for SdkInt
func (s *SdkInt) UnmarshalTOML(data []byte) error {
	var str string
	// Unmarshal the data as a string
	if err := toml.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("failed to unmarshal sdkmath.Int: %w", err)
	}

	// Convert the string to an sdkmath.Int
	res, ok := sdkmath.NewIntFromString(str)
	if !ok {
		return fmt.Errorf("invalid sdkmath.Int value: %s", str)
	}

	// Assign the result to the receiver
	s.Value = res
	return nil
}

// UnmarshalText implements TOML unmarshalling for SdkInt
func (s *SdkInt) UnmarshalText(text []byte) error {
	str := string(text) // Convert text to string
	res, ok := sdkmath.NewIntFromString(str)
	if !ok {
		return fmt.Errorf("invalid sdkmath.Int value: %s", str)
	}
	s.Value = res
	return nil
}

// MarshalText implements TOML marshalling for SdkInt
func (s SdkInt) MarshalText() ([]byte, error) {
	return []byte(s.Value.String()), nil
}

// DecodeHook for SdkInt
func SdkIntDecodeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if to != reflect.TypeOf(SdkInt{}) {
		return data, nil
	}

	switch from.Kind() {
	case reflect.String:
		str, ok := data.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for SdkInt, got %T", data)
		}
		value, ok := sdkmath.NewIntFromString(str)
		if !ok {
			return nil, fmt.Errorf("invalid sdkmath.Int value: %s", str)
		}
		return SdkInt{Value: value}, nil
	default:
		return nil, fmt.Errorf("unsupported type for SdkInt: %s", from.Kind())
	}
}
