package skipgo

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"go.uber.org/zap"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

const (
	// MaxPositionAmount6dp represents maximum allowed position value in USDC (10M)
	MaxPositionAmount6dp = 10_000_000
)

// PricePoint represents a point on the price curve
type PricePoint struct {
	Amount         *big.Int
	ExecutionPrice *big.Float
	PriceImpact    *big.Float
	Route          *RouteResponse
}

// FindOptimalSwapRoute finds the optimal amount to swap based on price impact
func (s *skipGoClient) FindOptimalSwapRoute(
	ctx context.Context,
	logger *zap.Logger,
	chainID, tokenIn, tokenOut string,
	maxAmount *big.Int,
	maxPriceImpact float64,
) (*big.Int, *RouteResponse, error) {
	// Define sample points to test (percentage of max amount)
	samplePoints := []float64{0.05, 0.1, 0.2, 0.3, 0.5, 0.7, 0.9, 1.0}

	var pricePoints []PricePoint
	var baselinePrice *big.Float

	// NOTE: this is biased to 6dp and needs improvement but dont get
	// too caught up on this
	// If the max amount is less than 10.000000, return the max amount
	if maxAmount.Cmp(big.NewInt(MaxPositionAmount6dp)) < 0 {
		// Get swap route for this amount
		route, err := s.SwapRoute(ctx, tokenIn, tokenOut, chainID, maxAmount)
		if err != nil {
			logger.Warn("Error getting swap route for test amount",
				zap.String("amount", maxAmount.String()),
				zap.Error(err),
			)
			return maxAmount, nil, err
		}

		return maxAmount, route, nil
	}

	// Sample the price curve at different amounts
	for i, percentage := range samplePoints {
		// Calculate test amount
		testAmount := new(big.Float).Mul(
			new(big.Float).SetInt(maxAmount),
			big.NewFloat(percentage),
		)

		// Convert to big.Int (rounding down)
		testAmountInt, _ := testAmount.Int(nil)

		// Skip if amount is zero
		if testAmountInt.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		// Get swap route for this amount
		route, err := s.SwapRoute(ctx, tokenIn, tokenOut, chainID, testAmountInt)
		if err != nil {
			logger.Warn("Error getting swap route for test amount",
				zap.String("amount", testAmountInt.String()),
				zap.Error(err),
			)
			continue
		}

		// Calculate execution price: amountIn / estimatedAmountOut
		amountIn := new(big.Float).SetInt(testAmountInt)
		estimatedAmountOut, _ := new(big.Float).SetString(route.EstimatedAmountOut)

		// Avoid division by zero
		if estimatedAmountOut.Cmp(big.NewFloat(0)) == 0 {
			logger.Warn("Estimated amount out is zero, skipping this test point",
				zap.String("amount", testAmountInt.String()),
			)
			continue
		}

		executionPrice := new(big.Float).Quo(amountIn, estimatedAmountOut)

		// Store the baseline price (from smallest amount)
		if i == 0 {
			baselinePrice = new(big.Float).Set(executionPrice)
		}

		// Calculate price impact compared to baseline
		priceImpact := new(big.Float).Sub(executionPrice, baselinePrice)
		priceImpact = new(big.Float).Quo(priceImpact, baselinePrice)
		priceImpact = new(big.Float).Mul(priceImpact, big.NewFloat(100)) // Convert to percentage

		pricePoints = append(pricePoints, PricePoint{
			Amount:         testAmountInt,
			ExecutionPrice: executionPrice,
			PriceImpact:    priceImpact,
			Route:          route,
		})

		logger.Debug("Price point sampled",
			zap.String("amount", testAmountInt.String()),
			zap.String("execution_price", executionPrice.Text('f', 8)),
			zap.String("price_impact", priceImpact.Text('f', 2)+"%"),
		)
	}

	// If we couldn't get any price points, return the max amount
	if len(pricePoints) == 0 {
		logger.Warn("Could not sample any price points, using max amount")

		// Try to get a route for the max amount
		route, err := s.SwapRoute(ctx, tokenIn, tokenOut, chainID, maxAmount)
		if err != nil {
			return maxAmount, nil, err
		}

		return maxAmount, route, nil
	}

	// Find the optimal price point
	var optimalPoint PricePoint
	maxAllowedImpact := big.NewFloat(maxPriceImpact)

	// Find the largest amount that has an impact below our threshold
	for i := len(pricePoints) - 1; i >= 0; i-- {
		point := pricePoints[i]

		// If price impact is below threshold, use this amount
		if point.PriceImpact.Cmp(maxAllowedImpact) <= 0 {
			optimalPoint = point
			logger.Info("Found optimal swap amount below impact threshold",
				zap.String("amount", point.Amount.String()),
				zap.String("price_impact", point.PriceImpact.Text('f', 2)+"%"),
			)
			return point.Amount, point.Route, nil
		}
	}

	// If no point is below threshold, use the one with lowest impact
	logger.Warn("No amount found below impact threshold, using amount with lowest impact")
	optimalPoint = pricePoints[0] // First point has lowest impact

	for _, point := range pricePoints {
		if point.PriceImpact.Cmp(optimalPoint.PriceImpact) < 0 {
			optimalPoint = point
		}
	}

	logger.Info("Selected optimal swap amount",
		zap.String("amount", optimalPoint.Amount.String()),
		zap.String("execution_price", optimalPoint.ExecutionPrice.Text('f', 8)),
		zap.String("price_impact", optimalPoint.PriceImpact.Text('f', 2)+"%"),
	)

	return optimalPoint.Amount, optimalPoint.Route, nil
}

// ProcessSwapEvent processes the incoming event, extracts the transaction hash, finds the matching recipient,
// and retrieves the associated amount.
func ProcessSwapEvent(events []abcitypes.Event, expectedSender, denom string) (string, error) {
	for _, event := range events {
		// Check if the event type is "transfer"
		if event.Type == "transfer" {
			var sender, amount string

			// Iterate through the attributes
			for _, attr := range event.Attributes {
				// Check for the "sender" attribute
				if attr.Key == "sender" && attr.Value == expectedSender {
					sender = attr.Value
				}

				// Check for the "amount" attribute with the specified denom suffix
				if attr.Key == "amount" && strings.HasSuffix(attr.Value, denom) {
					amount = attr.Value
				}
			}

			// If sender matches executor and amount is found, return the amount
			if sender == expectedSender && amount != "" {
				return amount, nil
			}
		}
	}

	return "", fmt.Errorf("no matching event found for expected sender: %s with denom: %s", expectedSender, denom)
}
