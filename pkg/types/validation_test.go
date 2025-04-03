package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config with Fees",
			config: Config{
				Chain: Chain{
					Prefix: "cosmos",
					Fees:   stringPtr("0.01uatom"),
					// Other fields...
				},
			},
			wantErr: false,
		},
		{
			name: "Valid config with Gas settings",
			config: Config{
				Chain: Chain{
					Prefix:        "cosmos",
					GasAdjustment: floatPtr(1.5),
					Gas:           stringPtr("200000"),
					GasPrices:     stringPtr("0.025uatom"),
					// Other fields...
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid config with both Fees and Gas settings",
			config: Config{
				Chain: Chain{
					Prefix:        "cosmos",
					Fees:          stringPtr("0.01uatom"),
					GasAdjustment: floatPtr(1.5),
					Gas:           stringPtr("200000"),
					GasPrices:     stringPtr("0.025uatom"),
				},
				// Other fields...
			},
			wantErr: true,
		},
		{
			name: "Invalid config with missing Gas settings",
			config: Config{
				Chain: Chain{
					Prefix:        "cosmos",
					GasAdjustment: floatPtr(1.5),
					Gas:           stringPtr("200000"),
					// Missing GasPrices
					// Other fields...
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid config with no Fees and Gas settings",
			config: Config{
				Chain: Chain{
					Prefix: "cosmos",
					// Missing Fees and Gas settings
					// Other fields...
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateConfig(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidConfig tests a valid configuration
func TestValidGridStrategyConfig(t *testing.T) {
	config := GridStrategyConfig{
		DefaultToken0Amount: SdkInt{Value: sdkmath.NewInt(100)},
		DefaultToken1Amount: SdkInt{Value: sdkmath.NewInt(200)},
		ChainID:             "cosmoshub-4",
		Granter:             "cosmos1...",
		Pool: Pool{
			ID:         1,
			BaseDenom:  "atom",
			QuoteDenom: "usdt",
		},
		Grid: Grid{
			Levels:     5,
			LowerBound: 10.5,
			UpperBound: 20.5,
		},
	}

	if err := config.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestInvalidConfig tests various invalid configurations
func TestInvalidGridStrategyConfig(t *testing.T) {
	tests := []struct {
		name   string
		config GridStrategyConfig
	}{
		{
			name: "missing DefaultToken0Amount",
			config: GridStrategyConfig{
				DefaultToken1Amount: SdkInt{Value: sdkmath.NewInt(200)},
				ChainID:             "cosmoshub-4",
				Granter:             "cosmos1...",
				Pool: Pool{
					ID:         1,
					BaseDenom:  "atom",
					QuoteDenom: "usdt",
				},
				Grid: Grid{
					Levels:     5,
					LowerBound: 10.5,
					UpperBound: 20.5,
				},
			},
		},
		{
			name: "empty ChainID",
			config: GridStrategyConfig{
				DefaultToken0Amount: SdkInt{Value: sdkmath.NewInt(100)},
				DefaultToken1Amount: SdkInt{Value: sdkmath.NewInt(200)},
				Granter:             "cosmos1...",
				Pool: Pool{
					ID:         1,
					BaseDenom:  "atom",
					QuoteDenom: "usdt",
				},
				Grid: Grid{
					Levels:     5,
					LowerBound: 10.5,
					UpperBound: 20.5,
				},
			},
		},
		{
			name: "invalid LowerBound",
			config: GridStrategyConfig{
				DefaultToken0Amount: SdkInt{Value: sdkmath.NewInt(100)},
				DefaultToken1Amount: SdkInt{Value: sdkmath.NewInt(200)},
				ChainID:             "cosmoshub-4",
				Granter:             "cosmos1...",
				Pool: Pool{
					ID:         1,
					BaseDenom:  "atom",
					QuoteDenom: "usdt",
				},
				Grid: Grid{
					Levels:     5,
					LowerBound: -10.5,
					UpperBound: 20.5,
				},
			},
		},
		// Add more invalid test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); err == nil {
				t.Errorf("expected error, got nil for test %s", tt.name)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}
