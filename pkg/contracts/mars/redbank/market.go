package redbank

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

// Market represents the market data structure.
type Market struct {
	Denom              string            `json:"denom"`
	ReserveFactor      string            `json:"reserve_factor"`
	InterestRateModel  InterestRateModel `json:"interest_rate_model"`
	BorrowIndex        string            `json:"borrow_index"`
	LiquidityIndex     string            `json:"liquidity_index"`
	BorrowRate         string            `json:"borrow_rate"`
	LiquidityRate      string            `json:"liquidity_rate"`
	IndexesLastUpdated uint64            `json:"indexes_last_updated"`
	CollateralTotal    string            `json:"collateral_total_scaled"`
	DebtTotal          string            `json:"debt_total_scaled"`
}

// NewMarket creates a new market with default values
func NewMarket(denom string) *Market {
	one := sdkmath.LegacyOneDec()
	zero := sdkmath.LegacyZeroDec()

	return &Market{
		Denom:         denom,
		ReserveFactor: zero.String(),
		InterestRateModel: InterestRateModel{
			OptimalUtilizationRate: zero.String(),
			Base:                   zero.String(),
			Slope1:                 zero.String(),
			Slope2:                 zero.String(),
		},
		BorrowIndex:        one.String(),
		LiquidityIndex:     one.String(),
		BorrowRate:         zero.String(),
		LiquidityRate:      zero.String(),
		IndexesLastUpdated: 0,
		CollateralTotal:    sdkmath.ZeroInt().String(),
		DebtTotal:          sdkmath.ZeroInt().String(),
	}
}

// Validate checks if the market parameters are valid
func (m *Market) Validate() error {
	one := sdkmath.LegacyOneDec()

	reserveFactor, err := sdkmath.LegacyNewDecFromStr(m.ReserveFactor)
	if err != nil {
		return fmt.Errorf("invalid reserve_factor: %s", m.ReserveFactor)
	}

	if reserveFactor.GT(one) {
		return fmt.Errorf("reserve_factor must be <= 1, got %v", reserveFactor)
	}

	// Convert string-based InterestRateModel to rational model for validation
	rationalModel, err := m.InterestRateModel.ToRational()
	if err != nil {
		return err
	}

	return rationalModel.Validate()
}

// UpdateInterestRates updates the borrow and liquidity rates based on the current utilization rate
func (m *Market) UpdateInterestRates(currentUtilizationRate sdkmath.LegacyDec) error {
	// Convert string-based InterestRateModel to rational model
	rationalModel, err := m.InterestRateModel.ToRational()
	if err != nil {
		return err
	}

	// Get reserve factor as decimal
	reserveFactor, err := sdkmath.LegacyNewDecFromStr(m.ReserveFactor)
	if err != nil {
		return fmt.Errorf("invalid reserve_factor: %s", m.ReserveFactor)
	}

	// Update borrow rate
	borrowRate, err := rationalModel.GetBorrowRate(currentUtilizationRate)
	if err != nil {
		return fmt.Errorf("failed to calculate borrow rate: %w", err)
	}

	// Update liquidity rate
	liquidityRate, err := rationalModel.GetLiquidityRate(
		borrowRate,
		currentUtilizationRate,
		reserveFactor,
	)
	if err != nil {
		return fmt.Errorf("failed to calculate liquidity rate: %w", err)
	}

	// Update the string values
	m.BorrowRate = borrowRate.String()
	m.LiquidityRate = liquidityRate.String()

	return nil
}

// IncreaseCollateral increases the total collateral by the given scaled amount
func (m *Market) IncreaseCollateral(amountScaled sdkmath.Int) error {
	currentCollateral, ok := sdkmath.NewIntFromString(m.CollateralTotal)
	if !ok {
		return fmt.Errorf("invalid collateral_total: %s", m.CollateralTotal)
	}

	newCollateral := currentCollateral.Add(amountScaled)
	m.CollateralTotal = newCollateral.String()
	return nil
}

// IncreaseDebt increases the total debt by the given scaled amount
func (m *Market) IncreaseDebt(amountScaled sdkmath.Int) error {
	currentDebt, ok := sdkmath.NewIntFromString(m.DebtTotal)
	if !ok {
		return fmt.Errorf("invalid debt_total: %s", m.DebtTotal)
	}

	newDebt := currentDebt.Add(amountScaled)
	m.DebtTotal = newDebt.String()
	return nil
}

// DecreaseCollateral decreases the total collateral by the given scaled amount
func (m *Market) DecreaseCollateral(amountScaled sdkmath.Int) error {
	currentCollateral, ok := sdkmath.NewIntFromString(m.CollateralTotal)
	if !ok {
		return fmt.Errorf("invalid collateral_total: %s", m.CollateralTotal)
	}

	if currentCollateral.LT(amountScaled) {
		return fmt.Errorf("insufficient collateral: %v < %v", currentCollateral, amountScaled)
	}

	newCollateral := currentCollateral.Sub(amountScaled)
	m.CollateralTotal = newCollateral.String()
	return nil
}

// DecreaseDebt decreases the total debt by the given scaled amount
func (m *Market) DecreaseDebt(amountScaled sdkmath.Int) error {
	currentDebt, ok := sdkmath.NewIntFromString(m.DebtTotal)
	if !ok {
		return fmt.Errorf("invalid debt_total: %s", m.DebtTotal)
	}

	if currentDebt.LT(amountScaled) {
		return fmt.Errorf("insufficient debt: %v < %v", currentDebt, amountScaled)
	}

	newDebt := currentDebt.Sub(amountScaled)
	m.DebtTotal = newDebt.String()
	return nil
}
