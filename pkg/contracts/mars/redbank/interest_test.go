package redbank

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeUnderlyingAmount(t *testing.T) {
	// Test with truncate operation
	scaledAmount := sdkmath.NewInt(1000000)
	index := sdkmath.LegacyMustNewDecFromStr("1.05")

	underlyingAmount, err := ComputeUnderlyingAmount(scaledAmount, index, Truncate)
	require.NoError(t, err)

	expectedAmount := sdkmath.NewInt(1050000)
	assert.Equal(t, expectedAmount, underlyingAmount)

	// Test with ceil operation
	underlyingAmount, err = ComputeUnderlyingAmount(scaledAmount, index, Ceil)
	require.NoError(t, err)

	// Should be the same in this case since the result is an integer
	assert.Equal(t, expectedAmount, underlyingAmount)

	// Test with a more complex decimal
	index = sdkmath.LegacyMustNewDecFromStr("1.0567")

	underlyingAmountTruncate, err := ComputeUnderlyingAmount(scaledAmount, index, Truncate)
	require.NoError(t, err)

	underlyingAmountCeil, err := ComputeUnderlyingAmount(scaledAmount, index, Ceil)
	require.NoError(t, err)

	// Ceil should be >= Truncate
	assert.True(t, underlyingAmountCeil.GTE(underlyingAmountTruncate))
}

func TestComputeScaledAmount(t *testing.T) {
	// Test with truncate operation
	underlyingAmount := sdkmath.NewInt(1050000)
	index := sdkmath.LegacyMustNewDecFromStr("1.05")

	scaledAmount, err := ComputeScaledAmount(underlyingAmount, index, Truncate)
	require.NoError(t, err)

	expectedAmount := sdkmath.NewInt(1000000)
	assert.Equal(t, expectedAmount, scaledAmount)

	// Test with a more complex decimal
	index = sdkmath.LegacyMustNewDecFromStr("1.0567")

	scaledAmountTruncate, err := ComputeScaledAmount(underlyingAmount, index, Truncate)
	require.NoError(t, err)

	scaledAmountCeil, err := ComputeScaledAmount(underlyingAmount, index, Ceil)
	require.NoError(t, err)

	// Ceil should be >= Truncate
	assert.True(t, scaledAmountCeil.GTE(scaledAmountTruncate))

	// Test division by zero
	_, err = ComputeScaledAmount(underlyingAmount, sdkmath.LegacyZeroDec(), Truncate)
	assert.Error(t, err)
}

func TestGetUnderlyingAmounts(t *testing.T) {
	// Create a market with some data
	market := NewMarket("uatom")
	market.LiquidityIndex = "1.05"
	market.BorrowIndex = "1.08"
	market.LiquidityRate = "0.03"
	market.BorrowRate = "0.05"
	market.IndexesLastUpdated = uint64(time.Now().Unix()) - 86400 // 1 day ago

	// Set some scaled amounts
	collateralScaled := sdkmath.NewInt(1000000)
	debtScaled := sdkmath.NewInt(700000)

	market.CollateralTotal = collateralScaled.String()
	market.DebtTotal = debtScaled.String()

	// Get current timestamp
	currentTimestamp := uint64(time.Now().Unix())

	// Test GetUnderlyingLiquidityAmount
	collateralAmount, err := GetUnderlyingLiquidityAmount(collateralScaled, market, currentTimestamp)
	require.NoError(t, err)

	// The index should have increased due to time elapsed
	assert.True(t, collateralAmount.GT(sdkmath.NewInt(1050000)))

	// Test GetUnderlyingDebtAmount
	debtAmount, err := GetUnderlyingDebtAmount(debtScaled, market, currentTimestamp)
	require.NoError(t, err)

	// The index should have increased due to time elapsed
	assert.True(t, debtAmount.GT(sdkmath.NewInt(756000))) // 700000 * 1.08

	// Test with zero scaled amounts
	zeroAmount, err := GetUnderlyingLiquidityAmount(sdkmath.ZeroInt(), market, currentTimestamp)
	require.NoError(t, err)
	assert.Equal(t, sdkmath.ZeroInt(), zeroAmount)

	zeroAmount, err = GetUnderlyingDebtAmount(sdkmath.ZeroInt(), market, currentTimestamp)
	require.NoError(t, err)
	assert.Equal(t, sdkmath.ZeroInt(), zeroAmount)
}

func TestCalculateUtilizationRate(t *testing.T) {
	// Create a market with some data
	market := NewMarket("uatom")
	market.LiquidityIndex = "1.0"
	market.BorrowIndex = "1.0"

	// Set some scaled amounts
	collateralScaled := sdkmath.NewInt(1000000)
	debtScaled := sdkmath.NewInt(700000)

	market.CollateralTotal = collateralScaled.String()
	market.DebtTotal = debtScaled.String()

	// Get current timestamp
	currentTimestamp := uint64(time.Now().Unix())

	// Test CalculateUtilizationRate
	utilizationRate, err := CalculateUtilizationRate(market, currentTimestamp)
	require.NoError(t, err)

	expectedRate := sdkmath.LegacyMustNewDecFromStr("0.7") // 700000/1000000
	assert.Equal(t, expectedRate, utilizationRate)

	// Test with zero collateral
	market.CollateralTotal = "0"
	utilizationRate, err = CalculateUtilizationRate(market, currentTimestamp)
	require.NoError(t, err)
	assert.Equal(t, sdkmath.LegacyZeroDec(), utilizationRate)

	// Test with debt > collateral (should cap at 1.0)
	market.CollateralTotal = "500000"
	market.DebtTotal = "700000"
	utilizationRate, err = CalculateUtilizationRate(market, currentTimestamp)
	require.NoError(t, err)
	assert.Equal(t, sdkmath.LegacyOneDec(), utilizationRate)
}
