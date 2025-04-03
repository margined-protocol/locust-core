package utils

import (
	"context"
	"errors"
	"math/big"

	"go.uber.org/zap"

	sdkmath "cosmossdk.io/math"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var ChainIDToPrefix = map[string]string{
	"osmosis-1":   "osmo",
	"neutron-1":   "neutron",
	"cosmoshub-4": "cosmos",
}

func CheckGas(ctx context.Context, l *zap.Logger, b banktypes.QueryClient, account, denom string) error {
	gasBalance, err := GetBalance(ctx, b, account, denom)
	if err != nil {
		l.Debug("Failed to get user balance", zap.Error(err))
		return err
	}

	if gasBalance.Balance.Amount.LTE(sdkmath.ZeroInt()) {
		l.Debug("User has no gas balance")
		return errors.New("No gas for: " + account + " in denom: " + denom)
	}

	return nil
}

// Contains checks if a slice contains a specific string.
func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func MinInt(x, y sdkmath.Int) sdkmath.Int {
	if x.LT(y) {
		return x
	}
	return y
}

func MinBigInt(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 { // x < y
		return x
	}
	return y
}
