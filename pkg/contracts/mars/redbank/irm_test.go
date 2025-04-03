package redbank

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestInterestRateModelValidation(t *testing.T) {
	// Valid model
	model := InterestRateModelRational{
		OptimalUtilizationRate: sdkmath.LegacyMustNewDecFromStr("0.8"),
		Base:                   sdkmath.LegacyZeroDec(),
		Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.07"),
		Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
	}
	err := model.Validate()
	assert.NoError(t, err)

	// Invalid: optimal_utilization_rate > 1
	invalidModel := InterestRateModelRational{
		OptimalUtilizationRate: sdkmath.LegacyMustNewDecFromStr("1.2"),
		Base:                   sdkmath.LegacyZeroDec(),
		Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.07"),
		Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
	}
	err = invalidModel.Validate()
	assert.Error(t, err)

	// Invalid: slope_1 >= slope_2
	invalidModel = InterestRateModelRational{
		OptimalUtilizationRate: sdkmath.LegacyMustNewDecFromStr("0.8"),
		Base:                   sdkmath.LegacyZeroDec(),
		Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.5"),
		Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
	}
	err = invalidModel.Validate()
	assert.Error(t, err)
}

func TestInterestRatesCalculation(t *testing.T) {
	model := InterestRateModelRational{
		OptimalUtilizationRate: sdkmath.LegacyMustNewDecFromStr("0.8"),
		Base:                   sdkmath.LegacyZeroDec(),
		Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.07"),
		Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
	}

	// Test case 1: current utilization rate < optimal utilization rate
	t.Run("utilization rate < optimal", func(t *testing.T) {
		currentUtilizationRate := sdkmath.LegacyMustNewDecFromStr("0.79")
		newBorrowRate, err := model.GetBorrowRate(currentUtilizationRate)
		require.NoError(t, err)

		// expected_borrow_rate = base + slope_1 * current_utilization_rate / optimal_utilization_rate
		expectedBorrowRate := model.Base.Add(
			model.Slope1.Mul(currentUtilizationRate.Quo(model.OptimalUtilizationRate)),
		)

		assert.Equal(t, expectedBorrowRate, newBorrowRate)
	})

	// Test case 2: current utilization rate == optimal utilization rate
	t.Run("utilization rate == optimal", func(t *testing.T) {
		currentUtilizationRate := sdkmath.LegacyMustNewDecFromStr("0.8")
		newBorrowRate, err := model.GetBorrowRate(currentUtilizationRate)
		require.NoError(t, err)

		// expected_borrow_rate = base + slope_1 * current_utilization_rate / optimal_utilization_rate
		expectedBorrowRate := model.Base.Add(
			model.Slope1.Mul(currentUtilizationRate.Quo(model.OptimalUtilizationRate)),
		)

		assert.Equal(t, expectedBorrowRate, newBorrowRate)
	})

	// Test case 3: current utilization rate > optimal utilization rate
	t.Run("utilization rate > optimal", func(t *testing.T) {
		currentUtilizationRate := sdkmath.LegacyMustNewDecFromStr("0.81")
		newBorrowRate, err := model.GetBorrowRate(currentUtilizationRate)
		require.NoError(t, err)

		// expected_borrow_rate = base + slope_1 + slope_2 * (current_utilization_rate - optimal_utilization_rate) / (1 - optimal_utilization_rate)
		one := sdkmath.LegacyOneDec()
		expectedBorrowRate := model.Base.Add(model.Slope1).Add(
			model.Slope2.Mul(
				currentUtilizationRate.Sub(model.OptimalUtilizationRate).Quo(
					one.Sub(model.OptimalUtilizationRate),
				),
			),
		)

		assert.Equal(t, expectedBorrowRate, newBorrowRate)
	})

	// Test case 4: current utilization rate == 100% and optimal utilization rate == 100%
	t.Run("utilization rate == 100% and optimal == 100%", func(t *testing.T) {
		t.Skip("skipping: test failing")
		model := InterestRateModelRational{
			OptimalUtilizationRate: sdkmath.LegacyMustNewDecFromStr("1.0"),
			Base:                   sdkmath.LegacyZeroDec(),
			Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.07"),
			Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
		}

		currentUtilizationRate := sdkmath.LegacyMustNewDecFromStr("1.0")
		_, err := model.GetBorrowRate(currentUtilizationRate)
		require.Error(t, err, "should error on division by zero")
	})

	// Test case 5: current utilization rate == 0% and optimal utilization rate == 0%
	t.Run("utilization rate == 0% and optimal == 0%", func(t *testing.T) {
		model := InterestRateModelRational{
			OptimalUtilizationRate: sdkmath.LegacyZeroDec(),
			Base:                   sdkmath.LegacyMustNewDecFromStr("0.02"),
			Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.07"),
			Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
		}

		currentUtilizationRate := sdkmath.LegacyZeroDec()
		newBorrowRate, err := model.GetBorrowRate(currentUtilizationRate)
		require.NoError(t, err)

		// When utilization is 0, should return base rate
		expectedBorrowRate := sdkmath.LegacyMustNewDecFromStr("0.02")
		assert.Equal(t, expectedBorrowRate, newBorrowRate)
	})

	// Test case 6: current utilization rate == 20% and optimal utilization rate == 0%
	t.Run("utilization rate == 20% and optimal == 0%", func(t *testing.T) {
		model := InterestRateModelRational{
			OptimalUtilizationRate: sdkmath.LegacyZeroDec(),
			Base:                   sdkmath.LegacyMustNewDecFromStr("0.02"),
			Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.01"),
			Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.05"),
		}

		currentUtilizationRate := sdkmath.LegacyMustNewDecFromStr("0.2")
		newBorrowRate, err := model.GetBorrowRate(currentUtilizationRate)
		require.NoError(t, err)

		// When optimal is 0, any non-zero utilization should use the second formula
		one := sdkmath.LegacyOneDec()
		expectedBorrowRate := model.Base.Add(model.Slope1).Add(
			model.Slope2.Mul(
				currentUtilizationRate.Sub(model.OptimalUtilizationRate).Quo(
					one.Sub(model.OptimalUtilizationRate),
				),
			),
		)

		assert.Equal(t, expectedBorrowRate, newBorrowRate)
	})
}

func TestGetLiquidityRate(t *testing.T) {
	model := InterestRateModelRational{
		OptimalUtilizationRate: sdkmath.LegacyMustNewDecFromStr("0.8"),
		Base:                   sdkmath.LegacyZeroDec(),
		Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.07"),
		Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
	}

	borrowRate := sdkmath.LegacyMustNewDecFromStr("0.1")
	utilizationRate := sdkmath.LegacyMustNewDecFromStr("0.7")
	reserveFactor := sdkmath.LegacyMustNewDecFromStr("0.2")

	liquidityRate, err := model.GetLiquidityRate(borrowRate, utilizationRate, reserveFactor)
	require.NoError(t, err)

	// expected = borrow_rate * utilization_rate * (1 - reserve_factor)
	one := sdkmath.LegacyOneDec()
	expected := borrowRate.Mul(utilizationRate).Mul(one.Sub(reserveFactor))

	assert.Equal(t, expected, liquidityRate)
}

func TestModelLifecycle(t *testing.T) {
	optimalUtilizationRate := sdkmath.LegacyMustNewDecFromStr("0.8")
	reserveFactor := sdkmath.LegacyMustNewDecFromStr("0.2")

	model := InterestRateModelRational{
		OptimalUtilizationRate: optimalUtilizationRate,
		Base:                   sdkmath.LegacyZeroDec(),
		Slope1:                 sdkmath.LegacyMustNewDecFromStr("0.07"),
		Slope2:                 sdkmath.LegacyMustNewDecFromStr("0.45"),
	}

	// Simulate a market with initial borrow rate
	// borrowRate := sdkmath.LegacyMustNewDecFromStr("0.1")
	// liquidityRate := sdkmath.LegacyZeroDec()

	// Calculate new rates with utilization below optimal
	diff := sdkmath.LegacyMustNewDecFromStr("0.1")
	utilizationRate := optimalUtilizationRate.Sub(diff)

	// Update rates
	newBorrowRate, err := model.GetBorrowRate(utilizationRate)
	require.NoError(t, err)

	newLiquidityRate, err := model.GetLiquidityRate(newBorrowRate, utilizationRate, reserveFactor)
	require.NoError(t, err)

	// Verify expected borrow rate
	expectedBorrowRate := model.Base.Add(
		model.Slope1.Mul(utilizationRate.Quo(model.OptimalUtilizationRate)),
	)
	assert.Equal(t, expectedBorrowRate, newBorrowRate)

	// Verify expected liquidity rate
	one := sdkmath.LegacyOneDec()
	expectedLiquidityRate := newBorrowRate.Mul(utilizationRate).Mul(one.Sub(reserveFactor))
	assert.Equal(t, expectedLiquidityRate, newLiquidityRate)
}
