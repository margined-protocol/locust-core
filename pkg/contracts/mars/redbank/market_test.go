package redbank

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestMarketValidation(t *testing.T) {
	// Valid market
	market := NewMarket("uatom")
	// nolint
	market.ReserveFactor = "0.2"
	market.InterestRateModel = InterestRateModel{
		OptimalUtilizationRate: "0.8",
		Base:                   "0",
		Slope1:                 "0.07",
		Slope2:                 "0.45",
	}

	err := market.Validate()
	assert.NoError(t, err)

	// Invalid: reserve_factor > 1
	invalidMarket := NewMarket("uatom")
	invalidMarket.ReserveFactor = "1.2"
	invalidMarket.InterestRateModel = InterestRateModel{
		OptimalUtilizationRate: "0.8",
		Base:                   "0",
		Slope1:                 "0.07",
		Slope2:                 "0.45",
	}

	err = invalidMarket.Validate()
	assert.Error(t, err)

	// Invalid: interest rate model validation fails
	invalidMarket = NewMarket("uatom")
	invalidMarket.ReserveFactor = "0.2"
	invalidMarket.InterestRateModel = InterestRateModel{
		OptimalUtilizationRate: "0.8",
		Base:                   "0",
		Slope1:                 "0.5",
		Slope2:                 "0.45",
	}

	err = invalidMarket.Validate()
	assert.Error(t, err)
}

func TestUpdateInterestRates(t *testing.T) {
	market := NewMarket("uatom")
	market.ReserveFactor = "0.2"
	market.InterestRateModel = InterestRateModel{
		OptimalUtilizationRate: "0.8",
		Base:                   "0",
		Slope1:                 "0.07",
		Slope2:                 "0.45",
	}

	// Test with utilization below optimal
	utilizationRate := sdkmath.LegacyMustNewDecFromStr("0.7")
	err := market.UpdateInterestRates(utilizationRate)
	require.NoError(t, err)

	// Get rational model for calculations
	rationalModel, err := market.InterestRateModel.ToRational()
	require.NoError(t, err)

	// Calculate expected borrow rate
	expectedBorrowRate := rationalModel.Base.Add(
		rationalModel.Slope1.Mul(
			utilizationRate.Quo(rationalModel.OptimalUtilizationRate),
		),
	)

	// Calculate expected liquidity rate
	one := sdkmath.LegacyOneDec()
	reserveFactor, err := sdkmath.LegacyNewDecFromStr(market.ReserveFactor)
	require.NoError(t, err)

	expectedLiquidityRate := expectedBorrowRate.Mul(utilizationRate).Mul(one.Sub(reserveFactor))

	assert.Equal(t, expectedBorrowRate.String(), market.BorrowRate)
	assert.Equal(t, expectedLiquidityRate.String(), market.LiquidityRate)

	// Test with utilization above optimal
	utilizationRate = sdkmath.LegacyMustNewDecFromStr("0.9")
	err = market.UpdateInterestRates(utilizationRate)
	require.NoError(t, err)

	// Calculate expected borrow rate for high utilization
	expectedBorrowRate = rationalModel.Base.Add(rationalModel.Slope1).Add(
		rationalModel.Slope2.Mul(
			utilizationRate.Sub(rationalModel.OptimalUtilizationRate).Quo(
				one.Sub(rationalModel.OptimalUtilizationRate),
			),
		),
	)

	// Calculate expected liquidity rate
	expectedLiquidityRate = expectedBorrowRate.Mul(utilizationRate).Mul(one.Sub(reserveFactor))

	assert.Equal(t, expectedBorrowRate.String(), market.BorrowRate)
	assert.Equal(t, expectedLiquidityRate.String(), market.LiquidityRate)
}

func TestCollateralAndDebtOperations(t *testing.T) {
	market := NewMarket("uatom")

	// Test increase collateral
	amount1 := sdkmath.NewInt(1000)
	err := market.IncreaseCollateral(amount1)
	require.NoError(t, err)
	assert.Equal(t, amount1.String(), market.CollateralTotal)

	// Test increase collateral again
	amount2 := sdkmath.NewInt(500)
	err = market.IncreaseCollateral(amount2)
	require.NoError(t, err)

	totalExpected := amount1.Add(amount2)
	assert.Equal(t, totalExpected.String(), market.CollateralTotal)

	// Test decrease collateral
	err = market.DecreaseCollateral(amount2)
	require.NoError(t, err)
	assert.Equal(t, amount1.String(), market.CollateralTotal)

	// Test decrease collateral too much
	err = market.DecreaseCollateral(amount1.Add(amount2))
	assert.Error(t, err)
	assert.Equal(t, amount1.String(), market.CollateralTotal)

	// Test increase debt
	err = market.IncreaseDebt(amount1)
	require.NoError(t, err)
	assert.Equal(t, amount1.String(), market.DebtTotal)

	// Test increase debt again
	err = market.IncreaseDebt(amount2)
	require.NoError(t, err)
	assert.Equal(t, totalExpected.String(), market.DebtTotal)

	// Test decrease debt
	err = market.DecreaseDebt(amount2)
	require.NoError(t, err)
	assert.Equal(t, amount1.String(), market.DebtTotal)

	// Test decrease debt too much
	err = market.DecreaseDebt(amount1.Add(amount2))
	assert.Error(t, err)
	assert.Equal(t, amount1.String(), market.DebtTotal)
}

func TestMarketLifecycle(t *testing.T) {
	// Create a market with interest rate model
	market := NewMarket("uatom")
	market.ReserveFactor = "0.2"
	market.InterestRateModel = InterestRateModel{
		OptimalUtilizationRate: "0.8",
		Base:                   "0",
		Slope1:                 "0.07",
		Slope2:                 "0.45",
	}

	// Add some collateral and debt
	collateralAmount := sdkmath.NewInt(10000)
	debtAmount := sdkmath.NewInt(7000)

	err := market.IncreaseCollateral(collateralAmount)
	require.NoError(t, err)

	err = market.IncreaseDebt(debtAmount)
	require.NoError(t, err)

	// Calculate utilization rate
	utilizationRate := sdkmath.LegacyNewDecFromInt(debtAmount).Quo(sdkmath.LegacyNewDecFromInt(collateralAmount))

	// Update interest rates
	err = market.UpdateInterestRates(utilizationRate)
	require.NoError(t, err)

	// Get rational model for calculations
	rationalModel, err := market.InterestRateModel.ToRational()
	require.NoError(t, err)

	// Verify rates are calculated correctly
	one := sdkmath.LegacyOneDec()
	reserveFactor, err := sdkmath.LegacyNewDecFromStr(market.ReserveFactor)
	require.NoError(t, err)

	// For utilization < optimal
	expectedBorrowRate := rationalModel.Base.Add(
		rationalModel.Slope1.Mul(
			utilizationRate.Quo(rationalModel.OptimalUtilizationRate),
		),
	)

	expectedLiquidityRate := expectedBorrowRate.Mul(utilizationRate).Mul(one.Sub(reserveFactor))

	assert.Equal(t, expectedBorrowRate.String(), market.BorrowRate)
	assert.Equal(t, expectedLiquidityRate.String(), market.LiquidityRate)

	// Simulate a withdrawal
	withdrawAmount := sdkmath.NewInt(2000)
	err = market.DecreaseCollateral(withdrawAmount)
	require.NoError(t, err)

	// Recalculate utilization rate
	newCollateralAmount := collateralAmount.Sub(withdrawAmount)
	utilizationRate = sdkmath.LegacyNewDecFromInt(debtAmount).Quo(
		sdkmath.LegacyNewDecFromInt(newCollateralAmount),
	)

	// Update interest rates with new utilization
	err = market.UpdateInterestRates(utilizationRate)
	require.NoError(t, err)

	// Verify rates are updated correctly
	if utilizationRate.LTE(rationalModel.OptimalUtilizationRate) {
		expectedBorrowRate = rationalModel.Base.Add(
			rationalModel.Slope1.Mul(
				utilizationRate.Quo(rationalModel.OptimalUtilizationRate),
			),
		)
	} else {
		expectedBorrowRate = rationalModel.Base.Add(rationalModel.Slope1).Add(
			rationalModel.Slope2.Mul(
				utilizationRate.Sub(rationalModel.OptimalUtilizationRate).Quo(
					one.Sub(rationalModel.OptimalUtilizationRate),
				),
			),
		)
	}

	expectedLiquidityRate = expectedBorrowRate.Mul(utilizationRate).Mul(one.Sub(reserveFactor))

	assert.Equal(t, expectedBorrowRate.String(), market.BorrowRate)
	assert.Equal(t, expectedLiquidityRate.String(), market.LiquidityRate)
}
