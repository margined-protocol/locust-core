package perps

import (
	"fmt"
	"strconv"

	"github.com/margined-protocol/locust-core/pkg/math"

	sdkmath "cosmossdk.io/math"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

func ProcessMarsPerpEvent(events []abcitypes.Event) (currentPrice string, entryPrice string, err error) {
	for _, event := range events {
		// Check if the event type is "wasm"
		if event.Type == "wasm" {
			var foundCurrentPrice, foundEntryPrice bool

			// Iterate through the attributes
			for _, attr := range event.Attributes {
				if attr.Key == "current_price" {
					currentPrice = attr.Value
					foundCurrentPrice = true
				}
				if attr.Key == "entry_price" {
					entryPrice = attr.Value
					foundEntryPrice = true
				}
			}

			// If current_price was found but entry_price was not, set entry_price = current_price
			if foundCurrentPrice && !foundEntryPrice {
				entryPrice = currentPrice
			}

			// If we found a current_price, return (entry_price will be set correctly)
			if foundCurrentPrice {
				return currentPrice, entryPrice, nil
			}
		}
	}

	return "", "", fmt.Errorf("no matching mars perp event found")
}

// ProcessCandlesResponse converts indexer response data into a Position
func ProcessCandlesResponse(
	_ string,
	quantumConversionExponent int64,
	_ int64,
	response *IndexerCandleResponse,
) (*sdkmath.Int, error) {
	if response == nil {
		return nil, fmt.Errorf("no candles found")
	}

	// Get the most recent candle
	candle := response.Candles[0] // First element is the most recent
	// candle := response.Candles[len(response.Candles)-1]

	// Convert the close price to a sdkmath.Int
	closePrice, err := strconv.ParseFloat(candle.Close, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse close price: %w", err)
	}

	// Convert the close price to a sdkmath.Int
	closePriceInt := math.FloatToQuantumPrice(closePrice, quantumConversionExponent)

	return &closePriceInt, nil
}

// ProcessIndexerResponse converts indexer response data into a Position
func ProcessIndexerResponse(
	market string,
	quantumConversionExponent int64,
	atomicResolution int64,
	response *IndexerSubaccountResponse,
) (*Position, error) {
	if response == nil {
		return &Position{
			EntryPrice:    sdkmath.ZeroInt(),
			Margin:        sdkmath.ZeroInt(),
			Amount:        sdkmath.ZeroInt(),
			CurrentPrice:  sdkmath.ZeroInt(),
			UnrealizedPnl: sdkmath.ZeroInt(),
			RealizedPnl:   sdkmath.ZeroInt(),
		}, nil
	}

	// Initialize position with zero values for all fields
	position := &Position{
		EntryPrice:    sdkmath.ZeroInt(),
		Margin:        sdkmath.ZeroInt(),
		Amount:        sdkmath.ZeroInt(),
		UnrealizedPnl: sdkmath.ZeroInt(),
		RealizedPnl:   sdkmath.ZeroInt(),
	}

	// Get margin from USDC asset position if it exists
	if assetPosition, hasAsset := response.Subaccount.AssetPositions["USDC"]; hasAsset {
		marginFloat, err := strconv.ParseFloat(assetPosition.Size, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse margin: %w", err)
		}
		position.Margin = math.FloatToQuantumPrice(marginFloat, atomicResolution)
	}

	// Get the position for our market
	perpPosition, exists := response.Subaccount.OpenPerpetualPositions[market]
	if !exists {
		return position, nil
	}

	// Convert entry price
	entryPriceFloat, err := strconv.ParseFloat(perpPosition.EntryPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse entry price: %w", err)
	}
	position.EntryPrice = math.FloatToQuantumPrice(entryPriceFloat, quantumConversionExponent)

	// Convert size
	sizeFloat, err := strconv.ParseFloat(perpPosition.Size, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse size: %w", err)
	}
	position.Amount = math.FloatToQuantumPrice(sizeFloat, atomicResolution)

	// Convert unrealized PnL
	unrealizedPnlFloat, err := strconv.ParseFloat(perpPosition.UnrealizedPnl, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unrealized PnL: %w", err)
	}
	position.UnrealizedPnl = math.FloatToQuantumPrice(unrealizedPnlFloat, quantumConversionExponent)

	// Convert realized PnL
	realizedPnlFloat, err := strconv.ParseFloat(perpPosition.RealizedPnl, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse realized PnL: %w", err)
	}
	position.RealizedPnl = math.FloatToQuantumPrice(realizedPnlFloat, quantumConversionExponent)

	return position, nil
}
