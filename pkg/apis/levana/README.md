
```go
	levanaAPIClient := levanaapi.NewClient(s.strategyConfig.LevanaAPIURL, httpClient)
```

```go
    // today
		endDate := time.Now().Format("2006-01-02")
    // yesterday
		startDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
		s.Logger.Debug("dates",
			zap.String("endDate", endDate),
			zap.String("startDate", startDate),
		)

    // fetch funding rates
		fundingRates, err := s.levanaAPIClient.FetchFundingRates(marketInfo.MarketAddr, startDate, endDate)
		if err != nil {
			s.Logger.Error("Error fetching funding rates")
			return nil, err
		}
		s.Logger.Debug("Fetched funding rates",
			zap.Any("rates", fundingRates),
		)

    // Filter the last 8 rates
		filteredRates := levanaapi.FilterRates(fundingRates, 8)
		s.Logger.Debug("filtered funding rates",
			zap.Any("rates", filteredRates),
		)

    // Estimate the next funding rate using SMA
		estLongRate, estShortRate, err := levanaapi.EstimateNextFundingRate(filteredRates)
		if err != nil {
			s.Logger.Error("Error calculating estimated rates")
			return nil, err
		}

		s.Logger.Debug("estimated rates",
			zap.Float64("long", estLongRate),
			zap.Float64("short", estShortRate),
		)

```
