package types

import (
	"errors"
)

// ValidateConfig checks if the configuration is valid.
func ValidateConfig(cfg Config) error {
	if cfg.Chain.Fees != nil && (cfg.Chain.GasAdjustment != nil || cfg.Chain.Gas != nil || cfg.Chain.GasPrices != nil) {
		return errors.New("if 'fees' is provided, 'gas', 'gas_adjustment', and 'gas_prices' must not be provided")
	}

	if cfg.Chain.Fees == nil && (cfg.Chain.GasAdjustment == nil || cfg.Chain.Gas == nil || cfg.Chain.GasPrices == nil) {
		return errors.New("'gas', 'gas_adjustment', and 'gas_prices' must all be provided together if 'fees' is not used")
	}

	return nil
}
