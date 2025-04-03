package pyth

import (
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"
)

// ConvertPythPrice converts the raw price using the exponent and returns a scaled osmomath.BigInt
func ConvertPythPrice(rawPrice string, exponent int) (sdkmath.Int, error) {
	// Convert raw price string to osmomath.BigInt
	priceBigInt, ok := sdkmath.NewIntFromString(rawPrice)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid price value: %s", rawPrice)
	}

	// Compute 10^abs(exponent) using osmomath Int
	scale := sdkmath.NewInt(int64(math.Pow10(int(math.Abs(float64(exponent))))))

	// Scale the price correctly based on the exponent
	if exponent < 0 {
		// Divide for negative exponent (scaling down)
		return priceBigInt.Quo(scale), nil
	}
	// Multiply for positive exponent (scaling up)
	return priceBigInt.Mul(scale), nil
}
