package levana

import (
	"fmt"
	"strconv"

	"github.com/cinar/indicator/v2/trend"
)

// FilterRates returns the last N funding rate entries
func FilterRates(rates []FundingRate, n int) []FundingRate {
	if len(rates) == 0 {
		return nil
	}

	// Ensure we don't request more entries than available
	if n > len(rates) {
		n = len(rates)
	}

	return rates[len(rates)-n:]
}

func ComputeEMA(rates []FundingRate, period int) (float64, float64, error) {
	if len(rates) < period {
		return 0, 0, fmt.Errorf("not enough data for EMA: got %d, need at least %d", len(rates), period)
	}

	longStream := make(chan float64)
	shortStream := make(chan float64)

	emaLong := trend.NewEmaWithPeriod[float64](period)
	emaShort := trend.NewEmaWithPeriod[float64](period)

	outLong := emaLong.Compute(longStream)
	outShort := emaShort.Compute(shortStream)

	// feed values into channels
	go func() {
		defer close(longStream)
		defer close(shortStream)

		for _, r := range rates {
			longVal, _ := strconv.ParseFloat(r.LongRate, 64)
			shortVal, _ := strconv.ParseFloat(r.ShortRate, 64)
			longStream <- longVal
			shortStream <- shortVal
		}
	}()

	var latestLong, latestShort float64
	for i := 0; i < len(rates)-emaLong.IdlePeriod(); i++ {
		latestLong = <-outLong
		latestShort = <-outShort
	}

	return latestLong, latestShort, nil
}

// EstimateNextFundingRate predicts the next funding rate using a simple moving average
func EstimateNextFundingRate(rates []FundingRate) (float64, float64, error) {
	if len(rates) == 0 {
		return 0, 0, nil
	}

	var longSum, shortSum float64
	var count int

	for _, rate := range rates {
		var longRate, shortRate float64

		// Check for errors when parsing
		if _, err := fmt.Sscanf(rate.LongRate, "%f", &longRate); err != nil {
			return 0, 0, err
		}

		if _, err := fmt.Sscanf(rate.ShortRate, "%f", &shortRate); err != nil {
			return 0, 0, err
		}

		longSum += longRate
		shortSum += shortRate
		count++
	}

	if count == 0 {
		return 0, 0, nil // Avoid division by zero
	}

	// Compute simple moving average (SMA)
	longAvg := longSum / float64(count)
	shortAvg := shortSum / float64(count)

	return longAvg, shortAvg, nil
}
