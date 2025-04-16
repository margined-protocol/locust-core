package helpers

import (
	"fmt"
)

func ErrFetchMarkets(err error) error {
	return fmt.Errorf("failed to fetch markets: %w", err)
}

func ErrFetchMarketInfo(marketID string, err error) error {
	return fmt.Errorf("market %s: failed to fetch market info: %w", marketID, err)
}

func ErrFetchMarketStatus(marketID string, err error) error {
	return fmt.Errorf("market %s: failed to fetch market status: %w", marketID, err)
}

func ErrFetchOpenPositions(marketAddr string, err error) error {
	return fmt.Errorf("market %s: failed to fetch open positions: %w", marketAddr, err)
}

func ErrFetchPositionDetails(marketAddr string, err error) error {
	return fmt.Errorf("market %s: failed to fetch position details: %w", marketAddr, err)
}
