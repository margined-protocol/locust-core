package astroport

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	sdkmath "cosmossdk.io/math"
)

// QueryClient is the API for querying an astroport contract.
type QueryClient interface {
	QuerySimulation(ctx context.Context, contractAddress, denom, amount string, opts ...grpc.CallOption) (*SimulationResponse, error)
	QueryPool(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*PoolResponse, error)
	BinarySearchHighestOfferAmount(ctx context.Context, contractAddress, denom string, targetPrice, precision float64) (float64, error)
	GetAvailableLiquidityAtPrice(ctx context.Context, contractAddress, denom string, targetPrice, precision float64) (float64, error)
	GetAvailableLiquidityInRange(ctx context.Context, contractAddress, denom string, buyPrice, sellPrice, precision float64) (float64, error)
	GetMaxLiquidityAtPrice(ctx context.Context, contractAddress, denom string, tagetPrice float64, logger *zap.Logger) (float64, error)
	GetAvailableLiquidityBelowTargetPrice(ctx context.Context, contractAddress, denom string, targetPrice, currentPrice, precision float64) (float64, error)
	MaximizeLiquidityInClosest25Percent(ctx context.Context, contractAddress, denom string, buyPrice, sellPrice, stepSize, precision float64) (float64, float64, error)
	MaximizeLiquidityAcrossSteps(ctx context.Context, contractAddress, denom string, buyPrice, sellPrice float64) (float64, sdkmath.Int, error)
	Close() error
}

type queryClient struct {
	baseQueryClient base.QueryClient
	cc              *grpc.ClientConn
}

var _ QueryClient = (*queryClient)(nil)

// NewQueryClient creates a new QueryClient
func NewQueryClient(conn *grpc.ClientConn) QueryClient {
	baseQueryClient := base.NewQueryClient(conn)

	return &queryClient{
		baseQueryClient: *baseQueryClient,
	}
}

// Close closes the gRPC connection to the server
func (q *queryClient) Close() error {
	return q.cc.Close()
}

func (q *queryClient) QuerySimulation(ctx context.Context, contractAddress, denom, amount string, opts ...grpc.CallOption) (*SimulationResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"simulation": map[string]any{
			"offer_asset": map[string]any{
				"info": map[string]any{
					"native_token": map[string]any{
						"denom": denom,
					},
				},
				"amount": amount,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var simulationResponse SimulationResponse
	if err := json.Unmarshal(rawResponseData, &simulationResponse); err != nil {
		return nil, err
	}

	return &simulationResponse, nil
}

func (q *queryClient) QueryPool(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*PoolResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{"pool": map[string]any{}})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var poolResponse PoolResponse
	if err := json.Unmarshal(rawResponseData, &poolResponse); err != nil {
		return nil, err
	}

	return &poolResponse, nil
}

// BinarySearchHighestOfferAmount performs a binary search to find the maximum offer amount to achieve a target price.
func (q *queryClient) BinarySearchHighestOfferAmount(ctx context.Context, contractAddress, denom string, targetPrice, precision float64) (float64, error) {
	// Initialize binary search bounds
	low := 0.0
	high := 1e12 // A high initial guess; adjust based on expected max offer

	for high-low > precision {
		mid := (low + high) / 2

		// Convert mid to an integer string without decimals
		midStr := fmt.Sprintf("%.0f", mid)

		// Query the simulation for the current mid-point offer amount
		simulationResult, err := q.QuerySimulation(ctx, contractAddress, denom, midStr)
		if err != nil {
			return 0, fmt.Errorf("error querying simulation: %w", err)
		}

		// Convert ReturnAmount to float64
		returnAmount, err := strconv.ParseFloat(simulationResult.ReturnAmount, 64)
		if err != nil {
			return 0, fmt.Errorf("error converting ReturnAmount to float64: %w", err)
		}

		// Calculate the price for the current offer amount
		currentPrice := returnAmount / mid

		if currentPrice >= targetPrice {
			// If the price is at least the desired price, increase the lower bound
			low = mid
		} else {
			// If the price is too low, decrease the upper bound
			high = mid
		}
	}
	// The maximum offer amount that gives at least the target price
	return low, nil
}

func (q *queryClient) GetAvailableLiquidityAtPrice(ctx context.Context, contractAddress, denom string, targetPrice, precision float64) (float64, error) {
	// Initialize binary search bounds
	low := 0.0
	high := 1e12
	totalLiquidity := 0.0

	for high-low > precision {
		mid := (low + high) / 2
		midStr := fmt.Sprintf("%.0f", mid)

		// Query the simulation for the current mid-point offer amount
		simulationResult, err := q.QuerySimulation(ctx, contractAddress, denom, midStr)
		if err != nil {
			return 0, fmt.Errorf("error querying simulation: %w", err)
		}

		returnAmount, err := strconv.ParseFloat(simulationResult.ReturnAmount, 64)
		if err != nil {
			return 0, fmt.Errorf("error converting ReturnAmount to float64: %w", err)
		}

		currentPrice := returnAmount / mid

		if currentPrice >= targetPrice {
			low = mid
			// Track liquidity at or below target price
			totalLiquidity += mid
		} else {
			high = mid
		}
	}

	// The total liquidity available at or below the target price
	return totalLiquidity, nil
}

// nolint
func (q *queryClient) GetAvailableLiquidityInRange(ctx context.Context, contractAddress, denom string, buyPrice, sellPrice, precision float64) (float64, error) {
	totalLiquidity := 0.0
	low := 1e3                             // Initial offer amount
	maxOfferAmount := 1e12                 // Upper limit for offer amounts
	var previousReturnAmount float64 = 0.0 // Track the liquidity already consumed

	// Loop to accumulate liquidity as the price increases from buyPrice to sellPrice
	for low <= maxOfferAmount {
		lowStr := fmt.Sprintf("%.0f", low)

		// Query the simulation for the current offer amount
		simulationResult, err := q.QuerySimulation(ctx, contractAddress, denom, lowStr)
		if err != nil {
			return 0, fmt.Errorf("error querying simulation: %w", err)
		}

		returnAmount, err := strconv.ParseFloat(simulationResult.ReturnAmount, 64)
		if err != nil {
			return 0, fmt.Errorf("error converting ReturnAmount to float64: %w", err)
		}

		// Calculate the current price
		currentPrice := returnAmount / low
		fmt.Printf("Offer amount: %s, Return amount: %f, Current price: %f\n", lowStr, returnAmount, currentPrice)

		// If price is within the profitable range, accumulate only the new liquidity (incremental change)
		if currentPrice >= buyPrice && currentPrice <= sellPrice {
			incrementalLiquidity := returnAmount - previousReturnAmount // New liquidity only
			if incrementalLiquidity > 0 {
				totalLiquidity += incrementalLiquidity
			}
			previousReturnAmount = returnAmount // Update the previously consumed liquidity
		} else if currentPrice > sellPrice {
			break // Stop if the price exceeds the sell price
		}

		// Increment offer amount to probe progressively worse prices
		low += low * 0.1 // Increase by 10%, adjust as necessary
	}

	fmt.Printf("Total liquidity found between %.5f and %.5f: %.2f\n", buyPrice, sellPrice, totalLiquidity)
	return totalLiquidity, nil
}

// nolint
func (q *queryClient) MaximizeLiquidityInClosest25Percent(ctx context.Context, contractAddress, denom string, buyPrice, sellPrice, stepSize, precision float64) (float64, float64, error) {
	// Calculate the price range and limit the search to the lower 25% of the price range
	priceRange := sellPrice - buyPrice
	upperLimit := buyPrice + (priceRange * 0.25)

	var maxLiquidity float64 = 0.0
	var bestPrice float64 = 0.0
	currentOfferAmount := 1e3 // Starting offer amount
	currentPrice := buyPrice  // Ensure we start at the buy price
	previousLiquidity := 0.0  // Track the previous liquidity to stop when it drops

	fmt.Printf("Searching for liquidity between %.5f and %.5f (upper limit), but will allow exceeding if profitable\n", buyPrice, upperLimit)

	// Loop from the buy price up to the sell price, but prioritize liquidity and profit
	for currentPrice <= sellPrice {
		offerStr := fmt.Sprintf("%.0f", currentOfferAmount)

		// Query the simulation for the current offer amount
		simulationResult, err := q.QuerySimulation(ctx, contractAddress, denom, offerStr)
		if err != nil {
			return 0, 0, fmt.Errorf("error querying simulation: %w", err)
		}

		returnAmount, err := strconv.ParseFloat(simulationResult.ReturnAmount, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("error converting ReturnAmount to float64: %w", err)
		}

		// Calculate the current price
		currentPrice = returnAmount / currentOfferAmount
		fmt.Printf("Offer amount: %s, Return amount: %f, Current price: %f\n", offerStr, returnAmount, currentPrice)

		// Calculate liquidity at this price point
		liquidityAtPrice := returnAmount / currentOfferAmount
		fmt.Printf("Price: %.5f, Liquidity: %.2f\n", currentPrice, liquidityAtPrice)

		// Track the deepest liquidity
		if liquidityAtPrice > maxLiquidity {
			maxLiquidity = liquidityAtPrice
			bestPrice = currentPrice
			fmt.Printf("New best price: %.5f with liquidity: %.2f\n", bestPrice, maxLiquidity)
		}

		// Stop if we reach 25% below the best price
		if currentPrice < bestPrice*0.75 {
			fmt.Printf("Stopping: Current price (%.5f) dropped below 25%% of best price (%.5f)\n", currentPrice, bestPrice)
			break
		}

		// Stop if liquidity drops drastically
		if liquidityAtPrice < previousLiquidity*0.9 {
			fmt.Println("Stopping due to significant drop in liquidity.")
			break
		}

		previousLiquidity = liquidityAtPrice // Update previous liquidity for the next iteration

		// Increment offer amount to probe the next price point (smaller step size for more controlled increments)
		currentOfferAmount += currentOfferAmount * 0.1 // Adjust step size to control granularity

		// Stop when we reach the sell price
		if currentPrice >= sellPrice {
			break
		}
	}

	// Check if we found any liquidity
	if bestPrice == 0.0 && maxLiquidity == 0.0 {
		return 0, 0, fmt.Errorf("no liquidity found in the range")
	}

	fmt.Printf("Best price found: %.5f with deepest liquidity: %.2f\n", bestPrice, maxLiquidity)
	return bestPrice, maxLiquidity, nil
}

func (q *queryClient) MaximizeLiquidityAcrossSteps(
	ctx context.Context,
	contractAddress, denom string,
	buyPrice, sellPrice float64,
) (float64, sdkmath.Int, error) {
	if buyPrice >= sellPrice {
		return 0, sdkmath.Int{}, fmt.Errorf("buy price %.5f is not less than sell price %.5f, no profit possible", buyPrice, sellPrice)
	}

	maxLiquidity := sdkmath.ZeroInt()
	bestPrice := 0.0

	// Calculate price range and step size
	priceRange := sellPrice - buyPrice
	stepAmount := 0.005 * priceRange // Dynamic step: adjust for larger ranges

	fmt.Printf("Searching for profitable liquidity between buy price: %.5f and sell price: %.5f\n", buyPrice, sellPrice)

	currentPrice := buyPrice + stepAmount

	for currentPrice <= sellPrice {
		fmt.Printf("Checking price point: %.5f\n", currentPrice)

		// Query liquidity at the current price
		liquidityFloat, err := q.findMaxLiquidityAtPrice(ctx, contractAddress, denom, currentPrice)
		if err != nil {
			fmt.Printf("Error querying liquidity at price %.5f: %v\n", currentPrice, err)
			break
		}

		// Scale and convert liquidity to sdkmath.Int
		scaledLiquidity := liquidityFloat * 1e6
		liquidity := sdkmath.NewIntFromUint64(uint64(scaledLiquidity))

		fmt.Printf("Price: %.5f, Liquidity: %s\n", currentPrice, liquidity)

		// Track the highest liquidity
		if liquidity.GT(maxLiquidity) {
			maxLiquidity = liquidity
			bestPrice = currentPrice
			fmt.Printf("New best price: %.5f with liquidity: %s\n", bestPrice, maxLiquidity)
		} else {
			// If liquidity decreases, stop the search
			fmt.Printf("Liquidity decreasing, stopping search at price: %.5f\n", currentPrice)
			break
		}

		// Increment the price and avoid floating-point issues
		currentPrice = roundToPrecision(currentPrice+stepAmount, 5)
	}

	// Check if any liquidity was found
	if maxLiquidity.IsZero() {
		return 0, sdkmath.Int{}, fmt.Errorf("no profitable liquidity found in range %.5f to %.5f", buyPrice, sellPrice)
	}

	fmt.Printf("Best profitable price: %.5f with liquidity: %s\n", bestPrice, maxLiquidity)
	return bestPrice, maxLiquidity, nil
}

// nolint
func (q *queryClient) findMaxLiquidityAtPrice(ctx context.Context, contractAddress, denom string, targetPrice float64) (float64, error) {
	currentOfferAmount := 1e3 // Start with a small offer amount (e.g., 1e3)
	incrementFactor := 2.0    // Exponential growth factor
	var maxLiquidity float64 = 0.0
	var lastValidOfferAmount float64 = currentOfferAmount
	previousLiquidity := 0.0 // Track previous liquidity to detect when it decreases

	// Perform the search by exponentially increasing the offer amount
	for {
		currentOfferAmountStr := fmt.Sprintf("%.0f", currentOfferAmount)

		// Query the simulation for the current offer amount
		simulationResult, err := q.QuerySimulation(ctx, contractAddress, denom, currentOfferAmountStr)
		if err != nil {
			return 0, fmt.Errorf("error querying simulation: %w", err)
		}

		returnAmount, err := strconv.ParseFloat(simulationResult.ReturnAmount, 64)
		if err != nil {
			return 0, fmt.Errorf("error converting ReturnAmount to float64: %w", err)
		}

		// Calculate the current price for this offer amount
		currentPrice := returnAmount / currentOfferAmount

		fmt.Printf("Offer amount: %s, Return amount: %f, Current price: %f\n", currentOfferAmountStr, returnAmount, currentPrice)

		// Check if the current price is still within the profitable range
		if currentPrice < targetPrice {
			fmt.Printf("Price %.5f is below the target price %.5f. Stopping search.\n", currentPrice, targetPrice)
			break // Exit if price falls below target price
		}

		// Use returnAmount as liquidity
		liquidity := returnAmount
		if liquidity > maxLiquidity {
			maxLiquidity = liquidity
			lastValidOfferAmount = currentOfferAmount
			fmt.Printf("New best price: %.5f with liquidity: %.2f\n", currentPrice, maxLiquidity)
		}

		// Stop if liquidity decreases significantly
		if liquidity < previousLiquidity {
			fmt.Printf("Liquidity decreased, stopping search.\n")
			break
		}

		// Update previous liquidity and increase offer amount exponentially
		previousLiquidity = liquidity
		currentOfferAmount *= incrementFactor // Exponentially increase the offer amount
	}

	fmt.Printf("Max liquidity found: %.2f for offer amount: %.2f\n", maxLiquidity, lastValidOfferAmount)
	return maxLiquidity, nil
}

func (q *queryClient) GetAvailableLiquidityBelowTargetPrice(
	ctx context.Context,
	contractAddress, denom string,
	_, targetPrice, _ float64,
) (float64, error) {
	increment := 1e5 // Adjust this increment to suit your asset precision
	totalLiquidity := 0.0
	offerAmount := 0.0

	for {
		offerAmount += increment
		offerAmountStr := fmt.Sprintf("%.0f", offerAmount)

		// Ensure offerAmount is non-zero
		if offerAmount <= 0 {
			return totalLiquidity, fmt.Errorf("offer amount is zero, increment might be too small")
		}

		simulationResult, err := q.QuerySimulation(ctx, contractAddress, denom, offerAmountStr)
		if err != nil {
			return totalLiquidity, fmt.Errorf("error querying simulation: %w", err)
		}

		returnAmount, err := strconv.ParseFloat(simulationResult.ReturnAmount, 64)
		if err != nil {
			return totalLiquidity, fmt.Errorf("error converting ReturnAmount to float64: %w", err)
		}

		currentSimPrice := returnAmount / offerAmount
		if currentSimPrice > targetPrice {
			break
		}
		totalLiquidity += offerAmount
	}
	return totalLiquidity, nil
}

func (q *queryClient) GetMaxLiquidityAtPrice(
	ctx context.Context,
	contractAddress, denom string,
	targetPrice float64,
	logger *zap.Logger,
) (float64, error) {
	offerAmount := 1e5 // Starting with an initial amount, adjust based on expected liquidity range
	totalLiquidity := 0.0

	for {
		offerAmountStr := fmt.Sprintf("%.0f", offerAmount)

		// Query the simulation for the current offer amount
		simulationResult, err := q.QuerySimulation(ctx, contractAddress, denom, offerAmountStr)
		if err != nil {
			logger.Error("Error querying simulation", zap.String("offerAmount", offerAmountStr), zap.Error(err))
			return totalLiquidity, fmt.Errorf("error querying simulation: %w", err)
		}

		// Convert ReturnAmount to float64
		returnAmount, err := strconv.ParseFloat(simulationResult.ReturnAmount, 64)
		if err != nil {
			logger.Error("Error converting ReturnAmount", zap.String("returnAmount", simulationResult.ReturnAmount), zap.Error(err))
			return totalLiquidity, fmt.Errorf("error converting ReturnAmount to float64: %w", err)
		}

		// Calculate the effective price for this offer amount
		currentPrice := returnAmount / offerAmount

		// Log the current state of the offer amount and resulting price
		logger.Debug("Simulation step",
			zap.Float64("offerAmount", offerAmount),
			zap.Float64("returnAmount", returnAmount),
			zap.Float64("currentPrice", currentPrice),
			zap.Float64("targetPrice", targetPrice),
			zap.Float64("totalLiquidity", totalLiquidity),
		)

		if currentPrice < targetPrice {
			logger.Info("Price exceeded target, terminating search",
				zap.Float64("offerAmount", offerAmount),
				zap.Float64("currentPrice", currentPrice),
				zap.Float64("targetPrice", targetPrice),
				zap.Float64("totalLiquidity", totalLiquidity),
			)
			break
		}

		// Update total liquidity at or below the target price
		totalLiquidity = offerAmount

		// Increase the offer amount exponentially for the next iteration
		offerAmount *= 2
	}

	logger.Info("Maximum liquidity found",
		zap.Float64("targetPrice", targetPrice),
		zap.Float64("totalLiquidity", totalLiquidity),
	)
	return totalLiquidity, nil
}

// Helper function to round a float to a specific number of decimal places
func roundToPrecision(value float64, decimals int) float64 {
	scale := math.Pow(10, float64(decimals))
	return math.Round(value*scale) / scale
}
