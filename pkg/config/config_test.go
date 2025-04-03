package config

import (
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/margined-protocol/locust-core/pkg/types"
	"github.com/test-go/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestSdkIntUnmarshalTOML(t *testing.T) {
	tomlData := `
default_token_0_amount = "500000000000000000000"
default_token_1_amount = "300000000000000000000"
chain_id = "osmosis-1"
granter = "osmo1xyz..."
`
	filePath := "test_config.toml"
	err := os.WriteFile(filePath, []byte(tomlData), 0o600)
	require.NoError(t, err)
	defer os.Remove(filePath)

	type TestConfig struct {
		DefaultToken0Amount types.SdkInt `toml:"default_token_0_amount" mapstructure:"default_token_0_amount"`
		DefaultToken1Amount types.SdkInt `toml:"default_token_1_amount" mapstructure:"default_token_1_amount"`
		ChainID             string       `toml:"chain_id" mapstructure:"chain_id"`
		Granter             string       `toml:"granter" mapstructure:"granter"`
	}

	var cfg TestConfig
	_, err = toml.DecodeFile(filePath, &cfg)
	require.NoError(t, err)

	res0, _ := sdkmath.NewIntFromString("500000000000000000000")
	res1, _ := sdkmath.NewIntFromString("300000000000000000000")

	require.Equal(t, res0, cfg.DefaultToken0Amount.Value)
	require.Equal(t, res1, cfg.DefaultToken1Amount.Value)
	require.Equal(t, "osmosis-1", cfg.ChainID)
	require.Equal(t, "osmo1xyz...", cfg.Granter)
}

func TestLoadConfig(t *testing.T) {
	tomlData := `# The memo to be sent with the transaction
memo = "locust"

# The signer account
signer_account = "margined-redemption"

# The number of times to retry failed transactions
tx_retry_count = 3

# The retry delay for failed transactions (in milliseconds)
tx_retry_delay_ms = 500

websocket_path = "/websocket"

[chain]
gas                 = "auto"
gas_adjustment      = 1.3
gas_denom           = "uosmo"
gas_prices          = "0.025uosmo"
grpc_server_address = "osmosis-grpc.polkachu.com:12590"
prefix              = "osmo"
rpc_server_address  = "https://rpc-osmosis.margined.io:443"
# Fees to be sent with the transaction
# fees = "10000uosmo"

[key]
app_name = "osmosis"
backend  = "pass"
root_dir = "/home/go"

[[strategy]]
name = "redemption-stride"

[strategy.config]
buffer          = "500000"
chain_id        = "dydx-mainnet-1"
float           = "200000000"
lsp_name        = "stride"
max_swap_amount = "500000000000000000000"
pricing_url     = "https://stride-api.polkachu.com/Stride-Labs/stride/stakeibc/host_zone"
threshold       = "0.005"
vault           = "osmo1grs74ux62lukjdlfs9ll4myyulcyjcd4uzwpw3cgnzdygtsspp2se2ense"

[strategy.config.pool]
base_denom  = "ibc/980E82A9F8E7CA8CD480F4577E73682A6D3855A267D1831485D7EBEF0E7A6C2C"
id          = 1423
quote_denom = "ibc/831F0B1BBB1D08A2B75311892876D71565478C532967545476DF4C2D7492E48C"

[strategy.config.to_lsp]
buffer             = "500000000000000000"
channel            = "channel-326"
denom              = "ibc/980E82A9F8E7CA8CD480F4577E73682A6D3855A267D1831485D7EBEF0E7A6C2C"
destination_prefix = "stride"
port               = "transfer"
source_prefix      = "osmo"

[strategy.config.from_base]
buffer             = "500000000000000000"
channel            = "channel-3"
denom              = "adydx"
destination_prefix = "osmo"
port               = "transfer"
source_prefix      = "dydx"

[strategy.config.unbond_stride]
buffer    = "500000"
denom     = "stadydx"
host_zone = "dydx-mainnet-1"

[strategy.config.lsd_chain]
gas                 = "auto"
gas_adjustment      = 1.3
gas_denom           = "ustrd"
gas_prices          = "0.025ustrd"
grpc_server_address = "stride-grpc.polkachu.com:12290"
prefix              = "stride"
rpc_server_address  = "https://stride-rpc.publicnode.com:443"

[strategy.config.base_chain]
gas                 = "auto"
gas_adjustment      = 5
gas_denom           = "adydx"
gas_prices          = "0.03adydx"
grpc_server_address = "dydx-grpc.polkachu.com:23890"
prefix              = "dydx"
rpc_server_address  = "https://dydx-rpc.polkachu.com:443"
# fees = "10000uatom"
`

	// Create a temporary file to store the TOML configuration
	tmpFile, err := os.CreateTemp(t.TempDir(), "test_config_*.toml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	// Write the TOML data to the temporary file
	_, err = tmpFile.Write([]byte(tomlData))
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// Load the configuration
	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	// Validate the top-level config fields
	require.Equal(t, "osmo", cfg.Chain.Prefix)
	require.Equal(t, "uosmo", cfg.Chain.GasDenom)
	require.Equal(t, "osmosis-grpc.polkachu.com:12590", cfg.Chain.GRPCServerAddress)
	require.Equal(t, "https://rpc-osmosis.margined.io:443", cfg.Chain.RPCServerAddress)
	require.False(t, cfg.Chain.GRPCTLS)
	require.Equal(t, "locust", cfg.Memo)
	require.Equal(t, "/websocket", cfg.WebsocketPath)
	require.Equal(t, "margined-redemption", cfg.SignerAccount)
	require.Equal(t, 3, cfg.TxRetryCount)
	require.Equal(t, 500*time.Millisecond, cfg.TxRetryDelay)
	require.False(t, cfg.DryRun)
	require.False(t, cfg.PropTrade)

	// Validate the strategy configuration
	require.Len(t, cfg.Strategies, 1)
	strategy := cfg.Strategies[0]
	require.Equal(t, "redemption-stride", strategy.Name)

	redemptionConfig, ok := strategy.Config.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "stride", redemptionConfig["lsp_name"])
	require.Equal(t, "500000", redemptionConfig["buffer"])
	require.Equal(t, "500000000000000000000", redemptionConfig["max_swap_amount"])
	require.Equal(t, "0.005", redemptionConfig["threshold"])
	require.Equal(t, "dydx-mainnet-1", redemptionConfig["chain_id"])
	require.Equal(t, "osmo1grs74ux62lukjdlfs9ll4myyulcyjcd4uzwpw3cgnzdygtsspp2se2ense", redemptionConfig["vault"])
	require.Equal(t, "200000000", redemptionConfig["float"])
	require.Equal(t, "https://stride-api.polkachu.com/Stride-Labs/stride/stakeibc/host_zone", redemptionConfig["pricing_url"])
}
