package math

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"

	sdkmath "cosmossdk.io/math"
)

// AdjustAmount adjusts the given amount based on the provided adjustment factor.
func AdjustAmount(amount sdkmath.Int, factor float64) sdkmath.Int {
	valueFloat := float64(amount.Int64())
	resultFloat := valueFloat * factor
	return sdkmath.NewInt(int64(resultFloat))
}

func AdjustSlippageFloat64(price float64, slippage float64, isBuy bool) float64 {
	if isBuy {
		return price * (1.0 + slippage)
	}
	return price * (1.0 - slippage)
}

func BpsToDecimal(bps int64) float64 {
	return float64(bps) / 10000.0
}

// ComparePercentageChange calculates the percentage change and checks if it is significant.
func ComparePercentageChange(oldValue, newValue float64, threshold int64) (float64, bool) {
	var percentageChange float64

	floatThreshold := BpsToDecimal(threshold)

	if oldValue != 0 {
		percentageChange = (newValue - oldValue) / math.Abs(oldValue)
	} else {
		percentageChange = 1
	}

	isSignificant := math.Abs(percentageChange) >= floatThreshold
	return percentageChange, isSignificant
}

func GenerateDeviation(baseValue float64, rng *rand.Rand) int64 {
	accuracyInt := int64(baseValue) // Cast float64 to int64
	deviation := rng.Int63n(accuracyInt*2) - accuracyInt
	return deviation
}

// interpolate linearly interpolates between two points (x1,y1) and (x2,y2)
// to find the y-value corresponding to the provided x-value
func Interpolate(
	x, x1, y1, x2, y2 sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	// If x2 == x1, return y1 to avoid division by zero
	if x2.Equal(x1) {
		return y1
	}

	// Linear interpolation formula: y = y1 + (x - x1) * (y2 - y1) / (x2 - x1)
	return y1.Add(
		x.Sub(x1).Mul(
			y2.Sub(y1),
		).Quo(
			x2.Sub(x1),
		),
	)
}

// RoundToNearestTickSpacing rounds a value to the nearest multiple of r.
func RoundToNearestTickSpacing(value, r int64) (int64, error) {
	if r <= 0 {
		return 0, fmt.Errorf("spacing (r) must be greater than 0")
	}

	// Calculate the nearest multiple of r
	rounded := math.Round(float64(value)/float64(r)) * float64(r)

	return int64(rounded), nil
}

// ParseCoinString converts a coin string to a big.Float value based on decimals
func ParseCoinString(token string, denom string, decimals int) (*big.Float, error) {
	if !strings.HasSuffix(token, denom) {
		return nil, fmt.Errorf("token %s does not contain expected denom %s", token, denom)
	}

	// Remove the denom from the token string
	valueStr := strings.TrimSuffix(token, denom)

	// Parse the numeric part as big.Int
	value := new(big.Int)
	if _, ok := value.SetString(valueStr, 10); !ok {
		return nil, fmt.Errorf("failed to parse token amount: %s", valueStr)
	}

	// Adjust for decimals
	adjustedValue := new(big.Float).SetInt(value)
	decimalFactor := new(big.Float).SetInt(pow10(decimals))
	adjustedValue.Quo(adjustedValue, decimalFactor)

	return adjustedValue, nil
}

// Convert float64 to sdkmath.LegacyDec
func FloatToLegacyDec(value float64) sdkmath.LegacyDec {
	strValue := fmt.Sprintf("%.18f", value) // Convert to string with 18 decimal places
	return sdkmath.LegacyMustNewDecFromStr(strValue)
}

// FloatToQuantumPrice converts a floating point price to a quantum-adjusted fixed-point integer
// Example: floatToQuantumPrice(1.234567899, -9) returns sdkmath.NewInt(1234567899)
func FloatToFixedInt(price float64, decimals int64) sdkmath.Int {
	// Convert the negative exponent to a positive multiplier
	// e.g., 6 becomes 1_000_000
	multiplier := math.Pow10(int(decimals))

	// Multiply the price by the multiplier and round to nearest integer
	fixedInt := math.Round(price * multiplier)

	// Convert to sdkmath.Int
	return sdkmath.NewInt(int64(fixedInt))
}

// FloatToQuantumPrice converts a floating point price to a quantum-adjusted fixed-point integer
// Example: floatToQuantumPrice(1.234567899, -9) returns sdkmath.NewInt(1234567899)
func FloatToQuantumPrice(price float64, quantumConversionExponent int64) sdkmath.Int {
	// Convert the negative exponent to a positive multiplier
	// e.g., -9 becomes 1_000_000_000
	multiplier := math.Pow10(int(-quantumConversionExponent))

	// Multiply the price by the multiplier and round to nearest integer
	quantumPrice := math.Round(price * multiplier)

	// Convert to sdkmath.Int
	return sdkmath.NewInt(int64(quantumPrice))
}

func RoundFixedPointInt(value sdkmath.Int, roundTo uint64) sdkmath.Int {
	roundToInt := sdkmath.NewInt(int64(roundTo))

	if roundToInt.IsZero() {
		panic("roundTo cannot be zero")
	}

	// Calculate quotient and remainder
	quotient := value.Quo(roundToInt)  // Integer division
	remainder := value.Mod(roundToInt) // Remainder

	// Get half of roundToInt for comparison
	halfRoundTo := roundToInt.Quo(sdkmath.NewInt(2))

	// If remainder >= halfRoundTo, round up
	if remainder.GTE(halfRoundTo) {
		return quotient.Add(sdkmath.OneInt()).Mul(roundToInt)
	}

	// Otherwise round down
	return quotient.Mul(roundToInt)
}

// Helper function for power of 10 using big.Int
func pow10(exp int) *big.Int {
	result := big.NewInt(1)
	ten := big.NewInt(10)
	for i := 0; i < exp; i++ {
		result.Mul(result, ten)
	}
	return result
}

// CalculatePriceFromCoins computes base/quote as a big.Float
func CalculatePriceFromCoins(base, baseDenom string, baseDecimals int, quote, quoteDenom string, quoteDecimals int) (*big.Float, error) {
	baseValue, err := ParseCoinString(base, baseDenom, baseDecimals)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base token: %w", err)
	}

	quoteValue, err := ParseCoinString(quote, quoteDenom, quoteDecimals)
	if err != nil {
		return nil, fmt.Errorf("failed to parse quote token: %w", err)
	}

	if quoteValue.Cmp(big.NewFloat(0)) == 0 {
		return nil, fmt.Errorf("quote token value is zero")
	}

	// Calculate price = quoteValue / baseValue
	price := new(big.Float).Quo(quoteValue, baseValue)

	return price, nil
}

// DivideWithDecimals divides a big.Int by a float64 considering the decimals
// and returns the result as a fixed-point big.Int with the specified decimals.
func DivideWithDecimals(f float64, b *big.Int, decimals int) *big.Int {
	scaleFactor := new(big.Float).SetFloat64(math.Pow(10, float64(decimals)))

	// Convert big.Int to big.Float
	bigB := new(big.Float).SetInt(b)

	bigB.Quo(bigB, scaleFactor)

	// Convert float64 to big.Float
	bigF := new(big.Float).SetFloat64(f)

	// Perform the division
	divided := new(big.Float).Quo(bigB, bigF)

	// Adjust for decimals
	divided.Mul(divided, scaleFactor)

	// Convert the result back to big.Int
	result := new(big.Int)
	divided.Int(result)

	return result
}

// MultiplyWithDecimals multiplies a float64 and a big.Int considering the decimals
// and returns the result as a fixed-point big.Int with the specified decimals.
func MultiplyWithDecimals(f float64, b *big.Int) *big.Int {
	// Convert float64 to big.Float
	bigF := new(big.Float).SetFloat64(f)

	// Convert big.Int to big.Float
	bigB := new(big.Float).SetInt(b)

	// Multiply the float and the big.Int
	multiplied := new(big.Float).Mul(bigF, bigB)

	result := new(big.Int)
	multiplied.Int(result)

	return result
}

// ConvertDecimalsBigInt truncates a fixed-point decimal from `fromDecimals` to `toDecimals`
func ConvertDecimalsBigInt(value *big.Int, fromDecimals, toDecimals int) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}

	exponent := toDecimals - fromDecimals

	// Compute 10^|exponent|
	scaleFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(exponent))), nil)

	scaledValue := new(big.Int).Set(value) // Copy input value

	if exponent > 0 {
		// Multiply for positive exponents
		scaledValue.Mul(scaledValue, scaleFactor)
	} else if exponent < 0 {
		// Divide for negative exponents
		scaledValue.Div(scaledValue, scaleFactor)
	}

	return scaledValue
}

// ConvertDecimalsSDK truncates sdkmath.Int from `fromDecimals` to `toDecimals`
func ConvertDecimalsSDK(value sdkmath.Int, fromDecimals, toDecimals int) sdkmath.Int {
	if fromDecimals == toDecimals {
		return value // No scaling needed
	}

	exponent := toDecimals - fromDecimals
	scaleFactor := sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(exponent))), nil))

	if exponent > 0 {
		// Multiply for positive exponent
		return value.Mul(scaleFactor)
	}

	// Divide for negative exponent (integer division)
	return value.Quo(scaleFactor)
}

// Helper function to get absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// SaturatingSub subtracts `subtrahend` from `minuend`, but ensures the result is never negative.
func SaturatingSub(minuend, subtrahend sdkmath.Int) sdkmath.Int {
	if minuend.LT(subtrahend) {
		return sdkmath.ZeroInt()
	}
	return minuend.Sub(subtrahend)
}
