package perps

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/margined-protocol/locust/core/pkg/math"
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
	market string,
	response *IndexerCandleResponse,
) (*sdkmath.LegacyDec, error) {
	if response == nil {
		return nil, fmt.Errorf("no candles found")
	}

	// Get the most recent candle
	candle := response.Candles[0] // First element is the most recent

	// Convert the close price to a sdkmath.Int
	closePrice, err := strconv.ParseFloat(candle.Close, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse close price: %w", err)
	}

	// Convert the close price to a sdkmath.Int
	closePriceInt := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", closePrice))

	return &closePriceInt, nil
}

// ProcessIndexerResponse converts indexer response data into a Position
func ProcessIndexerResponse(
	market string,
	decimals int64,
	response *IndexerSubaccountResponse,
) (*Position, error) {
	// Initialize position with zero values for all fields
	position := &Position{
		CurrentPrice:  sdkmath.LegacyZeroDec(),
		EntryPrice:    sdkmath.LegacyZeroDec(),
		Margin:        sdkmath.ZeroInt(),
		Amount:        sdkmath.ZeroInt(),
		UnrealizedPnl: sdkmath.ZeroInt(),
		RealizedPnl:   sdkmath.ZeroInt(),
	}

	if response == nil {
		return position, nil
	}

	// Equity is a proxy for Margin
	equity := response.Subaccount.Equity
	marginFloat, err := strconv.ParseFloat(equity, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse margin: %w", err)
	}
	position.Margin = math.FloatToFixedInt(marginFloat, decimals)

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
	position.EntryPrice = sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", entryPriceFloat))

	// Convert size
	sizeFloat, err := strconv.ParseFloat(perpPosition.Size, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse size: %w", err)
	}
	position.Amount = math.FloatToFixedInt(sizeFloat, decimals)

	// Convert unrealized PnL
	unrealizedPnlFloat, err := strconv.ParseFloat(perpPosition.UnrealizedPnl, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unrealized PnL: %w", err)
	}
	position.UnrealizedPnl = math.FloatToFixedInt(unrealizedPnlFloat, decimals)

	// Convert realized PnL
	realizedPnlFloat, err := strconv.ParseFloat(perpPosition.RealizedPnl, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse realized PnL: %w", err)
	}
	position.RealizedPnl = math.FloatToFixedInt(realizedPnlFloat, decimals)

	return position, nil
}
