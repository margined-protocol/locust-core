package fund

import (
	sdkmath "cosmossdk.io/math"
)

// CalculatePendingTotalShares sums up the shares across all pending redemptions
func CalculatePendingTotalShares(redemptions []Redemption) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, redemption := range redemptions {
		shares, ok := sdkmath.NewIntFromString(redemption.Amount)
		if !ok {
			continue // Skip invalid share values
		}
		total = total.Add(shares)
	}
	return total
}
