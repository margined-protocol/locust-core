package redbank

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
)

// ScalingOperation defines how rounding should be performed
type ScalingOperation int

const (
	// Truncate rounds down
	Truncate ScalingOperation = iota
	// Ceil rounds up
	Ceil
)

// GetUnderlyingLiquidityAmount calculates the actual liquidity amount from scaled amount
func GetUnderlyingLiquidityAmount(
	scaledAmount sdkmath.Int,
	market *Market,
	currentTimestamp uint64,
) (sdkmath.Int, error) {
	// If no scaled amount, return zero
	if scaledAmount.IsZero() {
		return sdkmath.ZeroInt(), nil
	}

	// Get liquidity index
	liquidityIndex, err := sdkmath.LegacyNewDecFromStr(market.LiquidityIndex)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("invalid liquidity_index: %s", market.LiquidityIndex)
	}

	// If indexes need updating, calculate the updated liquidity index
	if market.IndexesLastUpdated < currentTimestamp && !liquidityIndex.IsZero() {
		liquidityRate, err := sdkmath.LegacyNewDecFromStr(market.LiquidityRate)
		if err != nil {
			return sdkmath.Int{}, fmt.Errorf("invalid liquidity_rate: %s", market.LiquidityRate)
		}

		if !liquidityRate.IsZero() {
			timeElapsed := currentTimestamp - market.IndexesLastUpdated
			liquidityIndex, err = CalculateAppliedLinearInterestRate(
				liquidityIndex,
				liquidityRate,
				timeElapsed,
			)
			if err != nil {
				return sdkmath.Int{}, err
			}
		}
	}

	// Calculate underlying amount using the index
	return ComputeUnderlyingAmount(scaledAmount, liquidityIndex, Truncate)
}

// GetUnderlyingDebtAmount calculates the actual debt amount from scaled amount
func GetUnderlyingDebtAmount(
	scaledAmount sdkmath.Int,
	market *Market,
	currentTimestamp uint64,
) (sdkmath.Int, error) {
	// If no scaled amount, return zero
	if scaledAmount.IsZero() {
		return sdkmath.ZeroInt(), nil
	}

	// Get borrow index
	borrowIndex, err := sdkmath.LegacyNewDecFromStr(market.BorrowIndex)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("invalid borrow_index: %s", market.BorrowIndex)
	}

	// If indexes need updating, calculate the updated borrow index
	if market.IndexesLastUpdated < currentTimestamp && !borrowIndex.IsZero() {
		borrowRate, err := sdkmath.LegacyNewDecFromStr(market.BorrowRate)
		if err != nil {
			return sdkmath.Int{}, fmt.Errorf("invalid borrow_rate: %s", market.BorrowRate)
		}

		if !borrowRate.IsZero() {
			timeElapsed := currentTimestamp - market.IndexesLastUpdated
			borrowIndex, err = CalculateAppliedLinearInterestRate(
				borrowIndex,
				borrowRate,
				timeElapsed,
			)
			if err != nil {
				return sdkmath.Int{}, err
			}
		}
	}

	// Calculate underlying amount using the index
	return ComputeUnderlyingAmount(scaledAmount, borrowIndex, Ceil)
}

// ComputeUnderlyingAmount calculates the underlying amount from scaled amount and index
func ComputeUnderlyingAmount(
	scaledAmount sdkmath.Int,
	index sdkmath.LegacyDec,
	operation ScalingOperation,
) (sdkmath.Int, error) {
	// Convert scaled amount to decimal for calculation
	scaledAmountDec := sdkmath.LegacyNewDecFromInt(scaledAmount)

	// Calculate underlying amount: scaled_amount * index
	underlyingAmountDec := scaledAmountDec.Mul(index)

	// Apply rounding based on operation
	var underlyingAmount sdkmath.Int
	if operation == Truncate {
		underlyingAmount = underlyingAmountDec.TruncateInt()
	} else {
		// For Ceil operation, add a small amount before truncating to ensure rounding up
		// This is equivalent to ceiling the decimal value
		epsilon := sdkmath.LegacyNewDecWithPrec(1, 18) // Smallest possible decimal
		underlyingAmount = underlyingAmountDec.Add(epsilon).TruncateInt()
	}

	return underlyingAmount, nil
}

// ComputeScaledAmount calculates the scaled amount from underlying amount and index
func ComputeScaledAmount(
	underlyingAmount sdkmath.Int,
	index sdkmath.LegacyDec,
	operation ScalingOperation,
) (sdkmath.Int, error) {
	// Check for division by zero
	if index.IsZero() {
		return sdkmath.Int{}, fmt.Errorf("division by zero: index cannot be zero")
	}

	// Convert underlying amount to decimal for calculation
	underlyingAmountDec := sdkmath.LegacyNewDecFromInt(underlyingAmount)

	// Calculate scaled amount: underlying_amount / index
	scaledAmountDec := underlyingAmountDec.Quo(index)

	// Apply rounding based on operation
	var scaledAmount sdkmath.Int
	if operation == Truncate {
		scaledAmount = scaledAmountDec.TruncateInt()
	} else {
		// For Ceil operation, add a small amount before truncating to ensure rounding up
		epsilon := sdkmath.LegacyNewDecWithPrec(1, 18) // Smallest possible decimal
		scaledAmount = scaledAmountDec.Add(epsilon).TruncateInt()
	}

	return scaledAmount, nil
}

// CalculateAppliedLinearInterestRate calculates the updated index based on interest rate and time elapsed
func CalculateAppliedLinearInterestRate(
	currentIndex sdkmath.LegacyDec,
	interestRate sdkmath.LegacyDec,
	timeElapsed uint64,
) (sdkmath.LegacyDec, error) {
	// Calculate interest: rate * time_elapsed / seconds_per_year
	secondsPerYear := uint64(31536000) // 365 days

	// Convert to float64 for calculation to avoid potential overflow
	rateFloat, err := strconv.ParseFloat(interestRate.String(), 64)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to parse interest rate: %w", err)
	}

	timeElapsedFloat := float64(timeElapsed)
	secondsPerYearFloat := float64(secondsPerYear)

	// Calculate interest factor: 1 + (rate * time_elapsed / seconds_per_year)
	interestFactor := 1.0 + (rateFloat * timeElapsedFloat / secondsPerYearFloat)

	// Convert back to Dec
	interestFactorDec, err := sdkmath.LegacyNewDecFromStr(fmt.Sprintf("%f", interestFactor))
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to convert interest factor: %w", err)
	}

	// Calculate new index: current_index * interest_factor
	newIndex := currentIndex.Mul(interestFactorDec)

	return newIndex, nil
}

// CalculateUtilizationRate computes the current utilization rate for a market
func CalculateUtilizationRate(market *Market, currentTimestamp uint64) (sdkmath.LegacyDec, error) {
	// Get total collateral and debt
	collateralTotal, err := GetTotalCollateral(market, currentTimestamp)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	debtTotal, err := GetTotalDebt(market, currentTimestamp)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	// If no collateral, utilization is zero
	if collateralTotal.IsZero() {
		return sdkmath.LegacyZeroDec(), nil
	}

	// Calculate utilization rate: debt / collateral
	utilizationRate := sdkmath.LegacyNewDecFromInt(debtTotal).Quo(sdkmath.LegacyNewDecFromInt(collateralTotal))

	// Limit utilization rate to 100%
	one := sdkmath.LegacyOneDec()
	if utilizationRate.GT(one) {
		utilizationRate = one
	}

	return utilizationRate, nil
}

// GetTotalCollateral returns the total underlying collateral in the market
func GetTotalCollateral(market *Market, currentTimestamp uint64) (sdkmath.Int, error) {
	collateralScaled, ok := sdkmath.NewIntFromString(market.CollateralTotal)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid collateral_total: %s", market.CollateralTotal)
	}

	return GetUnderlyingLiquidityAmount(collateralScaled, market, currentTimestamp)
}

// GetTotalDebt returns the total underlying debt in the market
func GetTotalDebt(market *Market, currentTimestamp uint64) (sdkmath.Int, error) {
	debtScaled, ok := sdkmath.NewIntFromString(market.DebtTotal)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid debt_total: %s", market.DebtTotal)
	}

	return GetUnderlyingDebtAmount(debtScaled, market, currentTimestamp)
}
