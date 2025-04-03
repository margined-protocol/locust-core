package redbank

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

// InterestRateModel represents the interest rate model parameters.
type InterestRateModel struct {
	OptimalUtilizationRate string `json:"optimal_utilization_rate"`
	Base                   string `json:"base"`
	Slope1                 string `json:"slope_1"`
	Slope2                 string `json:"slope_2"`
}

// ToRational converts the string values to sdkmath.LegacyDec for calculations
func (m *InterestRateModel) ToRational() (*InterestRateModelRational, error) {
	optimal, err := sdkmath.LegacyNewDecFromStr(m.OptimalUtilizationRate)
	if err != nil {
		return nil, fmt.Errorf("invalid optimal_utilization_rate: %s", m.OptimalUtilizationRate)
	}

	base, err := sdkmath.LegacyNewDecFromStr(m.Base)
	if err != nil {
		return nil, fmt.Errorf("invalid base: %s", m.Base)
	}

	slope1, err := sdkmath.LegacyNewDecFromStr(m.Slope1)
	if err != nil {
		return nil, fmt.Errorf("invalid slope_1: %s", m.Slope1)
	}

	slope2, err := sdkmath.LegacyNewDecFromStr(m.Slope2)
	if err != nil {
		return nil, fmt.Errorf("invalid slope_2: %s", m.Slope2)
	}

	return &InterestRateModelRational{
		OptimalUtilizationRate: optimal,
		Base:                   base,
		Slope1:                 slope1,
		Slope2:                 slope2,
	}, nil
}

type InterestRateModelRational struct {
	OptimalUtilizationRate sdkmath.LegacyDec
	Base                   sdkmath.LegacyDec
	Slope1                 sdkmath.LegacyDec
	Slope2                 sdkmath.LegacyDec
}

func (m *InterestRateModelRational) Validate() error {
	one := sdkmath.LegacyOneDec()
	if m.OptimalUtilizationRate.GT(one) {
		return fmt.Errorf("optimal_utilization_rate must be <= 1, got %v", m.OptimalUtilizationRate)
	}

	if m.Slope1.GTE(m.Slope2) {
		return fmt.Errorf("slope_1 must be < slope_2, got %v >= %v", m.Slope1, m.Slope2)
	}

	return nil
}

func (m *InterestRateModelRational) GetBorrowRate(currentUtilizationRate sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	zero := sdkmath.LegacyZeroDec()
	one := sdkmath.LegacyOneDec()

	if currentUtilizationRate.LTE(m.OptimalUtilizationRate) {
		if currentUtilizationRate.Equal(zero) {
			return m.Base, nil
		}

		utilRatio := currentUtilizationRate.Quo(m.OptimalUtilizationRate)
		slope1Component := m.Slope1.Mul(utilRatio)
		return m.Base.Add(slope1Component), nil
	}

	excess := currentUtilizationRate.Sub(m.OptimalUtilizationRate)
	denominator := one.Sub(m.OptimalUtilizationRate)
	if denominator.Equal(zero) {
		return sdkmath.LegacyDec{}, fmt.Errorf("division by zero: optimal utilization rate cannot be 1")
	}

	slope2Component := m.Slope2.Mul(excess.Quo(denominator))
	result := m.Base.Add(m.Slope1).Add(slope2Component)

	return result, nil
}

func (m *InterestRateModelRational) GetLiquidityRate(borrowRate, currentUtilizationRate, reserveFactor sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	one := sdkmath.LegacyOneDec()

	reserveMultiplier := one.Sub(reserveFactor)
	result := borrowRate.Mul(currentUtilizationRate).Mul(reserveMultiplier)

	return result, nil
}
