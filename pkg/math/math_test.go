package math

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/test-go/testify/assert"

	sdkmath "cosmossdk.io/math"
)

// TestInterpolate tests the interpolation function used in interest rate calculations
func TestInterpolate(t *testing.T) {
	testCases := []struct {
		name     string
		x        string
		x1       string
		y1       string
		x2       string
		y2       string
		expected string
	}{
		{
			name:     "Middle point",
			x:        "3.0",
			x1:       "3.0",
			y1:       "11.1",
			x2:       "6.0",
			y2:       "17.4",
			expected: "11.1",
		},
		{
			name:     "Middle point",
			x:        "0.5",
			x1:       "0.0",
			y1:       "0.0",
			x2:       "1.0",
			y2:       "1.0",
			expected: "0.5",
		},
		{
			name:     "Equal x values should return y1",
			x:        "0.5",
			x1:       "0.5",
			y1:       "0.3",
			x2:       "0.5",
			y2:       "0.7",
			expected: "0.3",
		},
		{
			name:     "Interest rate kink interpolation",
			x:        "0.5",
			x1:       "0.0",
			y1:       "0.02",
			x2:       "0.8",
			y2:       "0.22",
			expected: "0.145",
		},
		{
			name:     "Interest rate above kink",
			x:        "0.85",
			x1:       "0.8",
			y1:       "0.22",
			x2:       "0.9",
			y2:       "1.52",
			expected: "0.87",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse values
			x, _ := sdkmath.LegacyNewDecFromStr(tc.x)
			x1, _ := sdkmath.LegacyNewDecFromStr(tc.x1)
			y1, _ := sdkmath.LegacyNewDecFromStr(tc.y1)
			x2, _ := sdkmath.LegacyNewDecFromStr(tc.x2)
			y2, _ := sdkmath.LegacyNewDecFromStr(tc.y2)
			expected, _ := sdkmath.LegacyNewDecFromStr(tc.expected)

			// Call interpolate function
			result := Interpolate(x, x1, y1, x2, y2)

			// Check with small tolerance for floating point errors
			delta := sdkmath.LegacyNewDecWithPrec(1, 6) // 0.000001 tolerance

			assert.InDelta(t,
				expected.MustFloat64(),
				result.MustFloat64(),
				delta.MustFloat64(),
				"Interpolation from (%s,%s) to (%s,%s) at x=%s should be %s, got %s",
				tc.x1, tc.y1, tc.x2, tc.y2, tc.x, tc.expected, result.String())
		})
	}
}

func TestRoundToNearestTickSpacing(t *testing.T) {
	tests := []struct {
		name      string
		value     int64
		r         int64
		want      int64
		expectErr bool
	}{
		{
			name:      "Round up to nearest r",
			value:     155,
			r:         100,
			want:      200,
			expectErr: false,
		},
		{
			name:      "Round down to nearest r",
			value:     149,
			r:         100,
			want:      100,
			expectErr: false,
		},
		{
			name:      "Value already aligned with r",
			value:     200,
			r:         100,
			want:      200,
			expectErr: false,
		},
		{
			name:      "Invalid r (zero spacing)",
			value:     200,
			r:         0,
			want:      0,
			expectErr: true,
		},
		{
			name:      "Negative r",
			value:     200,
			r:         -100,
			want:      0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RoundToNearestTickSpacing(tt.value, tt.r)
			if (err != nil) != tt.expectErr {
				t.Errorf("RoundToNearestTickSpacing() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoundToNearestTickSpacing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDivideWithDecimals(t *testing.T) {
	tests := []struct {
		name     string
		b        string  // big.Int value as string
		f        float64 // divisor
		decimals int
		expected string // expected result as string
	}{
		{
			name:     "Simple division",
			b:        "150",
			f:        1.5,
			decimals: 2,
			expected: "100",
		},
		{
			name:     "Small float, large int",
			b:        "250000",
			f:        0.25,
			decimals: 6,
			expected: "1000000",
		},
		{
			name:     "Large float, small int",
			b:        "123456",
			f:        12345.6789,
			decimals: 4,
			expected: "0009",
		},
		{
			name:     "Zero float",
			b:        "1000",
			f:        0.0,
			decimals: 3,
			expected: "0", // Division by zero should be handled appropriately
		},
		{
			name:     "Zero big.Int",
			b:        "0",
			f:        1.234,
			decimals: 3,
			expected: "0",
		},
		{
			name:     "Negative float",
			b:        "250",
			f:        -2.5,
			decimals: 2,
			expected: "-100",
		},
		{
			name:     "Negative big.Int",
			b:        "-250",
			f:        2.5,
			decimals: 2,
			expected: "-100",
		},
		{
			name:     "Both negative",
			b:        "-250",
			f:        -2.5,
			decimals: 2,
			expected: "100",
		},
		{
			name:     "No decimals",
			b:        "250",
			f:        2.5,
			decimals: 0,
			expected: "100",
		},
		{
			name:     "Worked example",
			b:        "577622",
			f:        5.77622199,
			decimals: 6,
			expected: "99999",
		},
		{
			name:     "18 decimals small values",
			b:        "5",
			f:        5.77622199,
			decimals: 18,
			expected: "0",
		},
		{
			name:     "18 decimals large values",
			b:        "5776221989999999806",
			f:        5.77622199,
			decimals: 18,
			expected: "999999999999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert input b and expected values to big.Int
			b := new(big.Int)
			b.SetString(tt.b, 10)

			expected := new(big.Int)
			expected.SetString(tt.expected, 10)

			// Call the function
			result := DivideWithDecimals(tt.f, b, tt.decimals)

			// Compare the result
			if result.Cmp(expected) != 0 {
				t.Errorf("DivideWithDecimals(%v, %v, %d) = %v; want %v",
					tt.b, tt.f, tt.decimals, result.String(), tt.expected)
			}
		})
	}
}

func TestMultiplyWithDecimals(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		b    string // big.Int value as string

		expected string // expected result as string
	}{
		{
			name:     "Simple multiplication",
			f:        1.5,
			b:        "100",
			expected: "150",
		},
		{
			name:     "Small float, large int",
			f:        0.25,
			b:        "1000000",
			expected: "250000",
		},
		{
			name:     "Large float, small int",
			f:        12345.6789,
			b:        "0010",
			expected: "123456",
		},
		{
			name:     "Zero float",
			f:        0.0,
			b:        "1000",
			expected: "0",
		},
		{
			name:     "Zero big.Int",
			f:        1.234,
			b:        "0",
			expected: "0",
		},
		{
			name:     "Negative float",
			f:        -2.5,
			b:        "100",
			expected: "-250",
		},
		{
			name:     "Negative big.Int",
			f:        2.5,
			b:        "-100",
			expected: "-250",
		},
		{
			name:     "Both negative",
			f:        -2.5,
			b:        "-100",
			expected: "250",
		},
		{
			name:     "No decimals",
			f:        2.5,
			b:        "100",
			expected: "250",
		},
		{
			name:     "Worked example",
			f:        5.77622199,
			b:        "100000",
			expected: "577622",
		},
		{
			name:     "18 decimals small values",
			f:        5.77622199,
			b:        "1",
			expected: "5",
		},
		{
			name:     "18 decimals small values",
			f:        5.77622199,
			b:        "1000000000000000000",
			expected: "5776221989999999806",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert input b and expected values to big.Int
			b := new(big.Int)
			b.SetString(tt.b, 10)

			expected := new(big.Int)
			expected.SetString(tt.expected, 10)

			// Call the function
			result := MultiplyWithDecimals(tt.f, b)

			// Compare the result
			if result.Cmp(expected) != 0 {
				t.Errorf("MultiplyWithDecimals(%v, %v) = %v; want %v",
					tt.f, tt.b, result.String(), tt.expected)
			}
		})
	}
}

// Test for big.Int
func TestConvertDecimalsBigInt(t *testing.T) {
	tests := []struct {
		input        string
		fromDecimals int
		toDecimals   int
		expected     string
	}{
		{"123456789012345678900000000000", 18, 6, "123456789012345678"},
		{"1000000000000000000", 18, 6, "1000000"},              // Expected conversion from 1*10^18 -> 1*10^6
		{"987654321098765432100000", 15, 6, "987654321098765"}, // Truncate last 9 digits
		{"123456789000000000000", 18, 6, "123456789"},          // Edge case: lower values
		{"1000000", 6, 18, "1000000000000000000"},              // Scale up from 6 to 18 decimals
		{"1", 6, 18, "1000000000000"},                          // Small value scale up
		{"1234567", 12, 12, "1234567"},
	}

	for _, test := range tests {
		inputBig := new(big.Int)
		inputBig.SetString(test.input, 10)
		expectedBig := new(big.Int)
		expectedBig.SetString(test.expected, 10)

		result := ConvertDecimalsBigInt(inputBig, test.fromDecimals, test.toDecimals)
		if result.Cmp(expectedBig) != 0 {
			t.Errorf("For input %s, expected %s but got %s", test.input, test.expected, result.String())
		}
	}
}

// Test for sdkmath.Int
func TestConvertDecimalsSDK(t *testing.T) {
	tests := []struct {
		input        string
		fromDecimals int
		toDecimals   int
		expected     string
	}{
		{"123456789012345678900000000000", 18, 6, "123456789012345678"},
		{"1000000000000000000", 18, 6, "1000000"},              // Expected conversion from 1*10^18 -> 1*10^6
		{"987654321098765432100000", 15, 6, "987654321098765"}, // Truncate last 9 digits
		{"123456789000000000000", 18, 6, "123456789"},          // Edge case: lower values
		{"1000000", 6, 18, "1000000000000000000"},              // Scale up from 6 to 18 decimals
		{"1", 6, 18, "1000000000000"},                          // Small value scale up
		{"1234567", 12, 12, "1234567"},
		{"-94973000000000000000000", 24, 6, "-94973"},
	}

	for _, test := range tests {
		// Convert input string to sdkmath.Int
		inputBigInt, success := new(big.Int).SetString(test.input, 10)
		if !success {
			t.Fatalf("Failed to parse input big.Int: %s", test.input)
		}
		inputSDK := sdkmath.NewIntFromBigInt(inputBigInt)

		// Convert expected string to sdkmath.Int
		expectedBigInt, success := new(big.Int).SetString(test.expected, 10)
		if !success {
			t.Fatalf("Failed to parse expected big.Int: %s", test.expected)
		}
		expectedSDK := sdkmath.NewIntFromBigInt(expectedBigInt)

		// Run the conversion function
		result := ConvertDecimalsSDK(inputSDK, test.fromDecimals, test.toDecimals)

		// Compare results
		if !result.Equal(expectedSDK) {
			t.Errorf("For input %s with decimals %d -> %d, expected %s but got %s",
				test.input, test.fromDecimals, test.toDecimals, test.expected, result.String())
		}
	}
}

func TestComparePercentageChange(t *testing.T) {
	tests := []struct {
		oldValue            float64
		newValue            float64
		threshold           int64
		expectedChange      float64
		expectedSignificant bool
	}{
		{35000.00, 36000.00, 250, 0.02857142857142857, true},
		{35000.00, 35500.00, 250, 0.014285714285714286, false},
		{0.00, 1000.00, 100, 1, true},
		{1000.00, 1000.00, 50, 0, false},
		{1000.00, 1050.00, 500, 0.05, true},
		{1000.00, 950.00, 450, -0.05, true},
		{0.0, 0, 1000, 1, true},
		{10000.0, 0, 1000, -1, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("oldValue: %v, newValue: %v", tt.oldValue, tt.newValue), func(t *testing.T) {
			change, significant := ComparePercentageChange(tt.oldValue, tt.newValue, tt.threshold)
			if change != tt.expectedChange {
				t.Errorf("expected change %.2f, got %.2f", tt.expectedChange, change)
			}
			if significant != tt.expectedSignificant {
				t.Errorf("expected significant %v, got %v", tt.expectedSignificant, significant)
			}
		})
	}
}

func TestFloatToQuantumPrice(t *testing.T) {
	tests := []struct {
		name                   string
		price                  float64
		quantumConversionExp   int64
		expectedQuantumizedInt int64
		expectedDecimalString  string
	}{
		{
			name:                   "simple whole number with -6 exponent",
			price:                  1.0,
			quantumConversionExp:   -6,
			expectedQuantumizedInt: 1_000_000,
			expectedDecimalString:  "1000000",
		},
		{
			name:                   "decimal number with -9 exponent",
			price:                  1.234567899,
			quantumConversionExp:   -9,
			expectedQuantumizedInt: 1_234_567_899,
			expectedDecimalString:  "1234567899",
		},
		{
			name:                   "small decimal with -9 exponent",
			price:                  0.000000001,
			quantumConversionExp:   -9,
			expectedQuantumizedInt: 1,
			expectedDecimalString:  "1",
		},
		{
			name:                   "rounding up with -6 exponent",
			price:                  1.5555555,
			quantumConversionExp:   -6,
			expectedQuantumizedInt: 1_555_556, // Should round up
			expectedDecimalString:  "1555556",
		},
		{
			name:                   "rounding down with -6 exponent",
			price:                  1.5555554,
			quantumConversionExp:   -6,
			expectedQuantumizedInt: 1_555_555, // Should round down
			expectedDecimalString:  "1555555",
		},
		{
			name:                   "zero price",
			price:                  0.0,
			quantumConversionExp:   -9,
			expectedQuantumizedInt: 0,
			expectedDecimalString:  "0",
		},
		{
			name:                   "large number with -6 exponent",
			price:                  1234567.89,
			quantumConversionExp:   -6,
			expectedQuantumizedInt: 1_234_567_890_000,
			expectedDecimalString:  "1234567890000",
		},
		{
			name:                   "different exponent (-3)",
			price:                  1.234,
			quantumConversionExp:   -3,
			expectedQuantumizedInt: 1_234,
			expectedDecimalString:  "1234",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FloatToQuantumPrice(tc.price, tc.quantumConversionExp)

			// Check the integer value
			require.Equal(t, tc.expectedQuantumizedInt, result.Int64(),
				"expected quantum price %d but got %d",
				tc.expectedQuantumizedInt, result.Int64())

			// Check string representation
			require.Equal(t, tc.expectedDecimalString, result.String(),
				"expected string representation %s but got %s",
				tc.expectedDecimalString, result.String())

			// Verify the result is a valid sdkmath.Int
			require.IsType(t, sdkmath.Int{}, result,
				"expected result to be sdkmath.Int type")
		})
	}
}

func TestRoundFixedPointInt(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		roundTo uint64
		want    int64
	}{
		{
			name:    "round up basic",
			value:   123456789,
			roundTo: 100,
			want:    123456800,
		},
		{
			name:    "round down basic",
			value:   123456749,
			roundTo: 100,
			want:    123456700,
		},
		{
			name:    "exact multiple",
			value:   123456700,
			roundTo: 100,
			want:    123456700,
		},
		{
			name:    "round up at exactly half",
			value:   123456750,
			roundTo: 100,
			want:    123456800,
		},
		{
			name:    "round with larger number",
			value:   123456789,
			roundTo: 1000,
			want:    123457000,
		},
		{
			name:    "round small number",
			value:   55,
			roundTo: 10,
			want:    60,
		},
		{
			name:    "zero value",
			value:   0,
			roundTo: 100,
			want:    0,
		},
		{
			name:    "negative round up",
			value:   -123456799,
			roundTo: 100,
			want:    -123456700,
		},
		{
			name:    "negative round down",
			value:   -123456749,
			roundTo: 100,
			want:    -123456600,
		},
		{
			name:    "large numbers",
			value:   1234567890123456789,
			roundTo: 1000000,
			want:    1234567890123000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := sdkmath.NewInt(tt.value)
			got := RoundFixedPointInt(value, tt.roundTo)

			require.Equal(t, tt.want, got.Int64(),
				"RoundFixedPointInt(%v, %v) = %v, want %v",
				tt.value, tt.roundTo, got, tt.want)

			// Verify result is multiple of roundTo
			require.Zero(t, got.Mod(sdkmath.NewInt(int64(tt.roundTo))).Int64(),
				"result should be multiple of roundTo")
		})
	}
}
