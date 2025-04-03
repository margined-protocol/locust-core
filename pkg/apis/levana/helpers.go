package levana

import "fmt"

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
