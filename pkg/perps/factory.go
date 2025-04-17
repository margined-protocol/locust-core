package perps

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/margined-protocol/locust-core/pkg/connection"
	"github.com/margined-protocol/locust-core/pkg/contracts/mars/creditmanager"
	marsperps "github.com/margined-protocol/locust-core/pkg/contracts/mars/perps"
	"github.com/margined-protocol/locust-core/pkg/ibc"
	subaccounts "github.com/margined-protocol/locust-core/pkg/proto/dydx/subaccounts/types"
	"github.com/margined-protocol/locust-core/pkg/types"
	"go.uber.org/zap"
)

// ProviderType represents supported perps providers
type ProviderType string

const (
	ProviderMars ProviderType = "mars"
	ProviderDydx ProviderType = "dydx"
	// Add other providers as needed
)

func GetProvider(provider string) ProviderType {
	switch provider {
	case "mars":
		return ProviderMars
	case "dydx":
		return ProviderDydx
	}
	return ProviderMars
}

// ProviderFactory creates perps providers
func CreateProvider(
	ctx context.Context,
	providerType ProviderType,
	logger *zap.Logger,
	config map[string]interface{},
) (Provider, error) {
	switch providerType {
	case ProviderMars:
		// Extract Mars-specific config and create Mars provider
		return createMarsProvider(logger, config)
	case ProviderDydx:
		// Create dYdX provider when implemented
		return createDydxProvider(logger, config)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// MarsConfig holds the configuration for the Mars provider
type MarsConfig struct {
	ChainID         string
	CreditClient    creditmanager.QueryClient
	PerpsClient     marsperps.QueryClient
	MarsConfig      types.MarsConfig
	CollateralDenom string
	OutDecimals     int
	Executor        string
}

// createMarsProvider creates a new MarsProvider from the provided configuration
func createMarsProvider(logger *zap.Logger, rawConfig map[string]interface{}) (Provider, error) {
	config, err := parseMarsConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Mars config: %w", err)
	}

	return NewMarsProvider(
		logger,
		config.ChainID,
		config.CreditClient,
		config.PerpsClient,
		config.MarsConfig,
		config.CollateralDenom,
		config.OutDecimals,
		config.Executor,
	), nil
}

// parseMarsConfig converts the raw config map into a strongly-typed MarsConfig
func parseMarsConfig(rawConfig map[string]interface{}) (MarsConfig, error) {
	var config MarsConfig
	var ok bool

	// Chain ID
	config.ChainID, ok = rawConfig["chain_id"].(string)
	if !ok {
		return config, fmt.Errorf("chain_id must be string")
	}

	// Credit Client
	config.CreditClient, ok = rawConfig["credit_client"].(creditmanager.QueryClient)
	if !ok {
		return config, fmt.Errorf("credit_client must be creditmanager.QueryClient")
	}

	// Perps Client
	config.PerpsClient, ok = rawConfig["perps_client"].(marsperps.QueryClient)
	if !ok {
		return config, fmt.Errorf("perps_client must be marsperps.QueryClient")
	}

	// Mars Config
	marsConfigRaw, ok := rawConfig["mars_config"].(types.MarsConfig)
	if !ok {
		return config, fmt.Errorf("mars_config must be types.MarsConfig")
	}
	config.MarsConfig = marsConfigRaw

	// Collateral Denom
	config.CollateralDenom, ok = rawConfig["collateral_denom"].(string)
	if !ok {
		return config, fmt.Errorf("collateral_denom must be string")
	}

	// Out Decimals
	outDecimals, ok := rawConfig["out_decimals"].(int)
	if !ok {
		return config, fmt.Errorf("out_decimals must be int")
	}
	config.OutDecimals = outDecimals

	// Executor
	config.Executor, ok = rawConfig["executor"].(string)
	if !ok {
		return config, fmt.Errorf("executor must be string")
	}

	return config, nil
}

// createDydxProvider creates a new DydxProvider from the provided configuration
func createDydxProvider(logger *zap.Logger, rawConfig map[string]interface{}) (Provider, error) {
	config, err := parseDydxConfig(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dYdX config: %w", err)
	}

	return NewDydxProvider(
		logger,
		config.MarketID,
		config.Market,
		config.SubticksPerTick,
		config.StepBaseQuantums,
		config.QuantumConversionExp,
		config.AtomicResolution,
		config.Decimals,
		config.MinEquity,
		config.SignerAccount,
		config.Executor,
		config.SubaccountID,
		config.Denom,
		config.BaseChainID,
		config.SubaccountClient,
		config.ClientRegistry,
		config.MsgHandler,
		config.IndexerURL,
	), nil
}

// parseDydxConfig converts the raw config map into a strongly-typed DydxConfig
func parseDydxConfig(rawConfig map[string]interface{}) (DydxConfig, error) {
	var config DydxConfig
	var ok bool

	// Required fields
	config.MarketID, ok = rawConfig["market_id"].(uint32)
	if !ok {
		return config, fmt.Errorf("market_id must be uint32")
	}

	config.Market, ok = rawConfig["market"].(string)
	if !ok {
		return config, fmt.Errorf("market must be string")
	}

	config.SubaccountID, ok = rawConfig["subaccount_id"].(uint32)
	if !ok {
		return config, fmt.Errorf("subaccount_id must be uint32")
	}

	config.SubaccountClient, ok = rawConfig["subaccount_client"].(subaccounts.QueryClient)
	if !ok {
		return config, fmt.Errorf("subaccount_client must be subaccounts.QueryClient")
	}

	config.Denom, ok = rawConfig["denom"].(string)
	if !ok {
		return config, fmt.Errorf("denom must be string")
	}

	config.BaseChainID, ok = rawConfig["base_chain_id"].(string)
	if !ok {
		return config, fmt.Errorf("base_chain_id must be string")
	}

	config.ClientRegistry, ok = rawConfig["client_registry"].(*connection.ClientRegistry)
	if !ok {
		return config, fmt.Errorf("client_registry must be *connection.ClientRegistry")
	}

	config.SignerAccount, ok = rawConfig["signer_account"].(string)
	if !ok {
		return config, fmt.Errorf("signer_account must be string")
	}

	config.MsgHandler, ok = rawConfig["msg_handler"].(ibc.MessageHandler)
	if !ok {
		return config, fmt.Errorf("msg_handler must be ibc.MessageHandler")
	}

	config.QuantumConversionExp, ok = rawConfig["quantum_conversion_exponent"].(int64)
	if !ok {
		return config, fmt.Errorf("quantum_conversion_exponent must be int64")
	}

	config.AtomicResolution, ok = rawConfig["atomic_resolution"].(int64)
	if !ok {
		return config, fmt.Errorf("atomic_resolution must be int64")
	}

	config.Decimals, ok = rawConfig["decimals"].(int64)
	if !ok {
		return config, fmt.Errorf("decimals must be int64")
	}

	config.SubticksPerTick, ok = rawConfig["subticks_per_tick"].(uint64)
	if !ok {
		return config, fmt.Errorf("subticks_per_tick must be uint64")
	}

	config.StepBaseQuantums, ok = rawConfig["step_base_quantums"].(uint64)
	if !ok {
		return config, fmt.Errorf("step_base_quantums must be uint64")
	}

	config.IndexerURL, ok = rawConfig["indexer_url"].(string)
	if !ok {
		return config, fmt.Errorf("indexer_url must be string")
	}

	config.Executor, ok = rawConfig["executor"].(string)
	if !ok {
		return config, fmt.Errorf("executor must be string")
	}

	config.MinEquity, ok = rawConfig["min_equity"].(sdkmath.Int)
	if !ok {
		return config, fmt.Errorf("min_equity must be sdkmath.Int")
	}

	return config, nil
}
