package helpers

import (
	"context"
	"sync"

	"github.com/margined-protocol/locust-core/pkg/contracts/levana"
	"github.com/margined-protocol/locust-core/pkg/contracts/levana/factory"
	"github.com/margined-protocol/locust-core/pkg/contracts/levana/market"
)

// FetchMarketsAndPositions fetches market info, status, and open positions for the given executor.
func FetchMarketsAndPositions(
	ctx context.Context,
	factoryClient factory.QueryClient,
	marketClient market.QueryClient,
	executor string,
) (map[string]levana.MarketData, error) {
	markets, err := factoryClient.Markets(ctx)
	if err != nil {
		return nil, ErrFetchMarkets(err)
	}

	marketDataMap := make(map[string]levana.MarketData)
	var mu sync.Mutex
	var wg sync.WaitGroup
	limit := 1000

	errCh := make(chan error, len(markets.Markets))

	for _, marketID := range markets.Markets {
		wg.Add(1)
		go func(marketID string) {
			defer wg.Done()

			marketInfo, err := factoryClient.MarketInfo(ctx, &factory.MarketInfoRequest{MarketID: marketID})
			if err != nil {
				errCh <- ErrFetchMarketInfo(marketID, err)
				return
			}

			status, err := marketClient.Status(ctx, marketInfo.MarketAddr)
			if err != nil {
				errCh <- ErrFetchMarketStatus(marketID, err)
				return
			}

			openPositions, err := marketClient.NftProxy(ctx, marketInfo.MarketAddr, executor, nil, &limit)
			if err != nil {
				errCh <- ErrFetchOpenPositions(marketInfo.MarketAddr, err)
				return
			}

			var positionsResponse *market.PositionsResponse
			if len(openPositions.Tokens) > 0 {
				positions, err := marketClient.Positions(ctx, marketInfo.MarketAddr, openPositions.Tokens)
				if err != nil {
					errCh <- ErrFetchPositionDetails(marketInfo.MarketAddr, err)
					return
				}
				positionsResponse = positions
			}

			mu.Lock()
			marketDataMap[marketInfo.MarketAddr] = levana.MarketData{
				MarketInfo: *marketInfo,
				Status:     *status,
				Positions:  positionsResponse,
			}
			mu.Unlock()
		}(marketID)
	}

	wg.Wait()
	close(errCh)

	// Return the first error (if any)
	for err := range errCh {
		return nil, err
	}

	return marketDataMap, nil
}
