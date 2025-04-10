package evaluator

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/cinar/indicator/v2/trend"
	levanaapi "github.com/margined-protocol/locust-core/pkg/apis/levana"
	"github.com/margined-protocol/locust-core/pkg/contracts/levana"
	"github.com/margined-protocol/locust-core/pkg/contracts/levana/factory"
	"github.com/margined-protocol/locust-core/pkg/contracts/levana/helpers"
	"github.com/margined-protocol/locust-core/pkg/contracts/levana/market"
	"github.com/margined-protocol/locust-core/pkg/contracts/pyth"

	sdkmath "cosmossdk.io/math"
)

type FundingRateEvaluator struct {
	mu                 sync.RWMutex
	decisions          map[string]levana.MarketDecision
	refreshEvery       time.Duration
	tradeSize          sdkmath.Int
	ctx                context.Context
	factoryQueryClient *factory.QueryClient
	marketQueryClient  *market.QueryClient
	pythClient         pyth.QueryClient
	levanaAPIClient    *levanaapi.Client
	marketDataList     map[string]levana.MarketData
	priceCache         map[string]float64
	executor           string
}

func NewFundingRateEvaluator(
	ctx context.Context,
	executor string,
	factoryQueryClient *factory.QueryClient,
	marketQueryClient *market.QueryClient,
	levanaAPIClient *levanaapi.Client,
	pythClient *pyth.QueryClient,
	refreshEvery time.Duration,
	tradeSize sdkmath.Int,
) *FundingRateEvaluator {
	return &FundingRateEvaluator{
		ctx:                ctx,
		executor:           executor,
		factoryQueryClient: factoryQueryClient,
		marketQueryClient:  marketQueryClient,
		levanaAPIClient:    levanaAPIClient,
		pythClient:         *pythClient,
		refreshEvery:       refreshEvery,
		tradeSize:          tradeSize,
		decisions:          make(map[string]levana.MarketDecision),
	}
}

func ComputeEMA(values []float64, period int) float64 {
	ema := trend.NewEmaWithPeriod[float64](period)
	input := make(chan float64, len(values))
	for _, v := range values {
		input <- v
	}
	close(input)

	var latest float64
	for v := range ema.Compute(input) {
		latest = v
	}
	return latest
}

func (e *FundingRateEvaluator) Start() {
	go e.refresh()
	ticker := time.NewTicker(e.refreshEvery)
	go func() {
		for {
			select {
			case <-e.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				e.refresh()
			}
		}
	}()
}

func (e *FundingRateEvaluator) refresh() {
	marketData, err := helpers.FetchMarketsAndPositions(e.ctx, *e.factoryQueryClient, *e.marketQueryClient, e.executor)
	if err != nil {
		return
	}

	e.mu.Lock()
	e.marketDataList = marketData
	e.mu.Unlock()

	tempDecisions := make(map[string]levana.MarketDecision)
	tempPriceCache := make(map[string]float64)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, data := range marketData {
		wg.Add(1)

		go func(data levana.MarketData) {
			defer wg.Done()

			status := data.Status
			config := status.Config

			// --- Fetch Pyth price ---
			if len(config.SpotPrice.Oracle.Feeds) > 0 && config.SpotPrice.Oracle.Feeds[0].Data.Pyth.ID != "" {
				pythID := config.SpotPrice.Oracle.Feeds[0].Data.Pyth.ID
				price, err := e.GetPythPriceScaled(pythID)
				if err == nil {
					mu.Lock()
					tempPriceCache[pythID] = price
					mu.Unlock()
				} else {
					// skip this market if price is essential
					return
				}
			} else {
				// skip market if no Pyth config
				return
			}

			// --- Fetch funding rates ---
			rates, err := e.levanaAPIClient.FetchFundingRates(
				data.MarketInfo.MarketAddr,
				time.Now().Add(-24*time.Hour).Format("2006-01-02"),
				time.Now().Format("2006-01-02"),
			)
			if err != nil {
				return
			}

			var shortRates, longRates []float64
			for _, r := range rates {
				s, err1 := strconv.ParseFloat(r.ShortRate, 64)
				l, err2 := strconv.ParseFloat(r.LongRate, 64)
				if err1 == nil {
					shortRates = append(shortRates, s)
				}
				if err2 == nil {
					longRates = append(longRates, l)
				}
			}
			if len(shortRates) < 24 || len(longRates) < 24 {
				return
			}

			emaShort := ComputeEMA(shortRates, 24)
			emaLong := ComputeEMA(longRates, 24)

			decisions := e.evaluateMarket(data, emaShort, emaLong)

			mu.Lock()
			for _, d := range decisions {
				key := d.MarketAddr + "|" + d.Direction
				tempDecisions[key] = d
			}
			mu.Unlock()
		}(data)
	}

	wg.Wait()

	e.mu.Lock()
	e.decisions = tempDecisions
	e.priceCache = tempPriceCache
	e.mu.Unlock()
}

func (e *FundingRateEvaluator) evaluateMarket(data levana.MarketData, emaShort, emaLong float64) []levana.MarketDecision {
	status := data.Status

	shortRate, _ := strconv.ParseFloat(status.ShortFunding, 64)
	longRate, _ := strconv.ParseFloat(status.LongFunding, 64)
	longUSD, _ := strconv.ParseFloat(status.LongUSD, 64)
	shortUSD, _ := strconv.ParseFloat(status.ShortUSD, 64)
	openInterest := longUSD + shortUSD
	shortNotional, _ := strconv.ParseFloat(status.ShortNotional, 64)
	longNotional, _ := strconv.ParseFloat(status.LongNotional, 64)
	imbalance := shortNotional / (shortNotional + longNotional + 1e-6)
	tradingFeeNotionalSize, _ := strconv.ParseFloat(status.Config.TradingFeeNotionalSize, 64)
	borrowFee, _ := strconv.ParseFloat(status.BorrowFee, 64)
	tradingFeeCounterCollateral, _ := strconv.ParseFloat(status.Config.TradingFeeCounterCollateral, 64)

	holdingHours := 48.0

	leverage := 3.0

	feeShort := EstimateFeePercentage(leverage, tradingFeeNotionalSize, tradingFeeCounterCollateral, borrowFee, shortRate, holdingHours)
	feeLong := EstimateFeePercentage(leverage, tradingFeeNotionalSize, tradingFeeCounterCollateral, borrowFee, longRate, holdingHours)

	projectedShort := ProjectFundingRate(shortRate, emaShort)
	projectedLong := ProjectFundingRate(longRate, emaLong)

	notionalTradeSize := e.tradeSize.ToLegacyDec().MustFloat64() * leverage

	// funding_rate_sensitivity
	fundingRateSensitivity, _ := strconv.ParseFloat(status.Config.FundingRateSensitivity, 64)

	// funding_rate_max_annualized
	fundingRateMaxAnnualized, _ := strconv.ParseFloat(status.Config.FundingRateMaxAnnualized, 64)

	// delta_neutrality_fee_sensitivity
	deltaNeutralityFeeSensitivity, _ := strconv.ParseFloat(status.Config.DeltaNeutralityFeeSensitivity, 64)

	// delta_neutrality_fee_cap
	deltaNeutralityFeeCap, _ := strconv.ParseFloat(status.Config.DeltaNeutralityFeeCap, 64)

	// Simulated trade for a short position (adding to shortNotional)
	newLongNotionalForShortTrade := longNotional
	newShortNotionalForShortTrade := shortNotional + notionalTradeSize
	_, projShortRateAfterShort := ComputeFundingRates(
		newLongNotionalForShortTrade, newShortNotionalForShortTrade,
		fundingRateSensitivity, fundingRateMaxAnnualized,
		deltaNeutralityFeeSensitivity, deltaNeutralityFeeCap,
	)
	shortProfit := EstimateProfit(projShortRateAfterShort, holdingHours, feeShort)

	// minProfitableLongRate := MinimumProfitableFundingRate(holdingHours, feeShort)

	// Simulated trade for a long position (adding to longNotional)
	newLongNotionalForLongTrade := longNotional + notionalTradeSize
	newShortNotionalForLongTrade := shortNotional
	projLongRateAfterLong, _ := ComputeFundingRates(
		newLongNotionalForLongTrade, newShortNotionalForLongTrade,
		fundingRateSensitivity, fundingRateMaxAnnualized,
		deltaNeutralityFeeSensitivity, deltaNeutralityFeeCap,
	)

	longProfit := EstimateProfit(projLongRateAfterLong, holdingHours, feeLong)

	// minProfitableShortRate := MinimumProfitableFundingRate(holdingHours, feeLong)

	now := time.Now()

	return []levana.MarketDecision{
		{
			MarketAddr:        data.MarketInfo.MarketAddr,
			MarketID:          data.Status.MarketID,
			MarketType:        data.Status.MarketType,
			Direction:         "short",
			ProfitEstimate:    shortProfit,
			CurrentRate:       shortRate,
			EMARate:           emaShort,
			OpenInterest:      openInterest,
			ProjectedPostRate: projectedShort,
			Fees:              feeShort,
			Imbalance:         imbalance,
			Action:            "ignore", // to be filled in later by the decision plan
			Timestamp:         now,
		},
		{
			MarketAddr:        data.MarketInfo.MarketAddr,
			MarketID:          data.Status.MarketID,
			MarketType:        data.Status.MarketType,
			Direction:         "long",
			ProfitEstimate:    longProfit,
			CurrentRate:       longRate,
			EMARate:           emaLong,
			OpenInterest:      openInterest,
			ProjectedPostRate: projectedLong,
			Fees:              feeLong,
			Imbalance:         imbalance,
			Action:            "ignore", // to be filled in later by the decision plan
			Timestamp:         now,
		},
	}
}

type DecisionView struct {
	Opportunities       []levana.MarketDecision
	PositionAdjustments []levana.MarketDecision
}

func (e *FundingRateEvaluator) GenerateDecisionPlan() DecisionView {
	view := DecisionView{
		Opportunities:       []levana.MarketDecision{},
		PositionAdjustments: []levana.MarketDecision{},
	}

	for _, marketData := range e.marketDataList {
		emaShort, emaLong := e.fetchEMA(marketData)
		pythID := marketData.Status.Config.SpotPrice.Oracle.Feeds[0].Data.Pyth.ID
		basePrice := e.priceCache[pythID]
		marketDecisions := e.evaluateMarket(marketData, emaShort, emaLong)

		// Evaluate existing positions if any
		if marketData.Positions != nil {
			for _, pos := range marketData.Positions.Positions {
				// TODO check if the funding rate has crossed
				// Use slices.IndexFunc to find and clone the corresponding market decision.
				idx := slices.IndexFunc(marketDecisions, func(d levana.MarketDecision) bool {
					return d.MarketID == marketData.Status.MarketID && d.Direction == pos.DirectionToBase
				})
				if idx == -1 {
					continue
				}
				// Clone the decision so we can use base data.
				positionDecision := marketDecisions[idx]
				percentDiff := (positionDecision.CurrentRate - positionDecision.EMARate) / math.Abs(positionDecision.EMARate)
				positionDecision.HasPosition = true
				positionID, _ := strconv.ParseInt(pos.ID, 10, 64)
				positionDecision.PositionID = positionID

				// Check if we should close this position
				shouldExit, reason := e.CheckPositionExit(pos, positionDecision.CurrentRate, basePrice)
				if shouldExit {
					positionDecision.Action = "close"
					positionDecision.Note = reason
					view.PositionAdjustments = append(view.PositionAdjustments, positionDecision)
					continue
				}

				// Funding signal logic: if funding rate deviation is significant.
				if percentDiff < -0.5 {
					positionDecision.Action = "increase"
					view.PositionAdjustments = append(view.PositionAdjustments, positionDecision)
					continue
				}
				if percentDiff > 0.5 {
					positionDecision.Action = "reduce"
					view.PositionAdjustments = append(view.PositionAdjustments, positionDecision)
					continue
				}

				// Otherwise, hold the position.
				positionDecision.Action = "hold"
				view.PositionAdjustments = append(view.PositionAdjustments, positionDecision)

			}
		}
		for _, decision := range marketDecisions {
			decision.Action = "ignore" // initially set

			// If there's no matching position, evaluate as opportunity
			if decision.Action == "ignore" && decision.ProfitEstimate > 0 {
				decision.Action = "open"
				view.Opportunities = append(view.Opportunities, decision)
			}
		}

	}

	sort.SliceStable(view.Opportunities, func(i, j int) bool {
		return view.Opportunities[i].ProfitEstimate > view.Opportunities[j].ProfitEstimate
	})

	sort.SliceStable(view.PositionAdjustments, func(i, j int) bool {
		return view.PositionAdjustments[i].ProfitEstimate > view.PositionAdjustments[j].ProfitEstimate
	})

	return view
}

func (e *FundingRateEvaluator) fetchEMA(data levana.MarketData) (emaShort, emaLong float64) {
	rates, err := e.levanaAPIClient.FetchFundingRates(
		data.MarketInfo.MarketAddr,
		time.Now().Add(-24*time.Hour).Format("2006-01-02"),
		time.Now().Format("2006-01-02"),
	)
	if err != nil {
		return 0, 0 // fallback
	}

	var shortRates, longRates []float64
	for _, r := range rates {
		s, err1 := strconv.ParseFloat(r.ShortRate, 64)
		l, err2 := strconv.ParseFloat(r.LongRate, 64)
		if err1 == nil {
			shortRates = append(shortRates, s)
		}
		if err2 == nil {
			longRates = append(longRates, l)
		}
	}

	if len(shortRates) < 24 || len(longRates) < 24 {
		return 0, 0
	}

	emaShort = ComputeEMA(shortRates, 24)
	emaLong = ComputeEMA(longRates, 24)

	return emaShort, emaLong
}

// CheckPositionExit assesses whether the position should be closed
// - is the position in range of liquidation?
// - is ihe position in range of take profit?
// - is the funding rate positive?
func (e *FundingRateEvaluator) CheckPositionExit(pos market.Position, currentRate, currentPrice float64) (shouldExit bool, reason string) {
	// Buffer for take profit and liquiditation prices
	const buffer = 0.05 // 5%

	// If the funding rate is no longer negative close
	if currentRate > 0 {
		return true, "funding-positive"
	}

	// Parse target prices from the position
	liquidationPrice, err1 := strconv.ParseFloat(pos.LiquidationPriceBase, 64)
	takeProfitPrice, err2 := strconv.ParseFloat(pos.TakeProfitPriceBase, 64)
	if err1 != nil || err2 != nil {
		return false, ""
	}

	switch pos.DirectionToBase {
	case "long":
		// For long positions, liquidation is below current price
		if liquidationPrice > 0 {
			relativeDiffLiquidation := (currentPrice - liquidationPrice) / liquidationPrice
			if relativeDiffLiquidation <= buffer {
				return true, "liquidation-risk"
			}
		}
		// For long positions, take profit is above current price
		if takeProfitPrice > 0 {
			relativeDiffTakeProfit := (takeProfitPrice - currentPrice) / takeProfitPrice
			if relativeDiffTakeProfit <= buffer {
				return true, "take-profit"
			}
		}
	case "short":
		// For short positions, liquidation is above current price
		if liquidationPrice > 0 {
			relativeDiffLiquidation := (liquidationPrice - currentPrice) / liquidationPrice
			if relativeDiffLiquidation <= buffer {
				return true, "liquidation-risk"
			}
		}
		// For short positions, take profit is below current price
		if takeProfitPrice > 0 {
			relativeDiffTakeProfit := (currentPrice - takeProfitPrice) / takeProfitPrice
			if relativeDiffTakeProfit <= buffer {
				return true, "take-profit"
			}
		}
	}
	// If no conditions are met, do not exit the position.
	return false, ""
}

// SuggestInvestment returns the next best market decision based on current capital.
func (e *FundingRateEvaluator) SuggestInvestment(view DecisionView) *levana.MarketDecision {
	if len(view.Opportunities) > 0 {
		if len(view.PositionAdjustments) == 0 || view.Opportunities[0].ProfitEstimate > view.PositionAdjustments[0].ProfitEstimate {
			return &view.Opportunities[0]
		}
	}

	if len(view.PositionAdjustments) > 0 && view.PositionAdjustments[0].ProfitEstimate > 0 {
		return &view.PositionAdjustments[0]
	}

	return nil
}

func CalcEntryFeePercent(
	leverage float64,
	feeRateNotional float64,
	feeRateCounter float64,
) float64 {
	// Fee_N = Notional * feeRateNotional → as % of deposit → same
	// Fee_G = Half of (Notional - Deposit) → = Notional * (1 - 1/Leverage) / 2
	// So:
	feeNotional := feeRateNotional
	feeCounter := ((1 - (1 / leverage)) / 2.0) * feeRateCounter

	return feeNotional + feeCounter
}

func (e *FundingRateEvaluator) GetPythPriceScaled(pythID string) (float64, error) {
	rawPythPrice, err := e.pythClient.LatestPrice(e.ctx, pythID)
	if err != nil {
		return 0, fmt.Errorf("error fetching Pyth price: %w", err)
	}

	if len(rawPythPrice.Parsed) == 0 || rawPythPrice.Parsed[0].Price.Price == "" {
		return 0, fmt.Errorf("invalid Pyth price response for ID: %s", pythID)
	}

	parsedPrice, err := strconv.ParseFloat(rawPythPrice.Parsed[0].Price.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing Pyth price: %w", err)
	}

	exponent := rawPythPrice.Parsed[0].Price.Exponent
	scalingFactor := math.Pow(10, math.Abs(float64(exponent)))

	var scaled float64
	if exponent < 0 {
		scaled = parsedPrice / scalingFactor
	} else {
		scaled = parsedPrice * scalingFactor
	}

	return scaled, nil
}

func EstimateFeePercentage(
	leverage float64,
	tradingFeeRateNotional float64, // FeeN, e.g. 0.0005 (0.05%)
	tradingFeeRateCounter float64, // FeeG, e.g. 0.001  (0.1%)
	borrowRateAnnualized float64, // e.g. 0.03
	fundingRateAnnualized float64, // e.g. -0.3669
	holdingHours float64,
) float64 {
	daysHeld := holdingHours / 24.0

	// 1. Trading Fee on notional
	tradingFeeNotional := tradingFeeRateNotional // 0.05% of notional

	// 2. Trading Fee on borrowed portion (counter collateral)
	// Borrowed portion = notional - deposit = deposit * (leverage - 1)
	// As % of deposit: (leverage - 1) * FeeG
	tradingFeeCounter := (leverage - 1) * tradingFeeRateCounter

	// 3. Borrow fee on borrowed portion
	borrowFee := (1 - 1/leverage) * (borrowRateAnnualized / 365.0) * daysHeld

	// 4. Funding fee on full notional
	fundingFee := fundingRateAnnualized / 365.0 * daysHeld

	// Total impact on collateral
	return tradingFeeNotional + tradingFeeCounter + borrowFee + fundingFee
}

func ProjectFundingRate(current float64, ema float64) float64 {
	delta := current - ema

	switch {
	case delta > 0.01:
		// Strongly positive funding → maybe cooling off
		return current - 0.015
	case delta > 0:
		// Mildly above EMA → slightly lower
		return current - 0.01
	case delta < -0.01:
		// Below EMA → could rebound
		return current + 0.01
	default:
		// Near EMA → hold steady
		return current
	}
}

// CompuateFundingRates computes funding rates for long and short markets
// For the formal definintion see https://docs.levana.finance/whitepaper#441-funding-rates
func ComputeFundingRates(
	longNotional float64,
	shortNotional float64,
	fundingRateSensitivity float64, // funding_rate_sensitivity
	fundingRateMaxAnnualized float64, // funding_rate_max_annualized
	deltaNeutralityFeeSensitivity float64, // delta_neutrality_fee_sensitivity (raw, needs scaling)
	deltaNeutralityFeeCap float64, // delta_neutrality_fee_cap
) (longRate float64, shortRate float64) {
	// Calculate total notional returning zero rates if both are zero
	totalNotional := longNotional + shortNotional
	if totalNotional == 0 {
		return 0, 0
	}

	// Calculate the net open interest
	netOpenInterest := longNotional - shortNotional

	// Determine the popular and unpopular sides
	var popularNotional, unpopularNotional float64
	if netOpenInterest > 0 {
		// Longs are more popular
		popularNotional = longNotional
		unpopularNotional = shortNotional
	} else {
		// Shorts are more popular
		popularNotional = shortNotional
		unpopularNotional = longNotional
	}

	// Compute effective funding rate sensitivityj
	fundingRateOverride := (fundingRateMaxAnnualized * (totalNotional / (deltaNeutralityFeeSensitivity * deltaNeutralityFeeCap)))
	effectiveSensitivity := math.Max(fundingRateOverride, fundingRateSensitivity)

	// Compute the raw popular funding rate (without capping).
	rawPopularRate := (netOpenInterest / totalNotional) * effectiveSensitivity

	popularFundingRate := math.Min(rawPopularRate, fundingRateMaxAnnualized)

	unpopularFundingRate := popularFundingRate * (popularNotional / unpopularNotional)
	if netOpenInterest > 0 {
		return popularFundingRate, -unpopularFundingRate
	}
	return unpopularFundingRate, -popularFundingRate
}

func EstimateProfit(
	fundingRateAnnualized float64, // from status.ShortFunding or LongFunding
	holdingHours float64,
	entryFee float64,
) float64 {
	// Convert annualized rate to rate over the holding period
	fundingImpact := -fundingRateAnnualized * (holdingHours / 8760.0)
	return fundingImpact - entryFee
}

func MinimumProfitableFundingRate(holdingHours, entryFee float64) float64 {
	return -entryFee * (8760.0 / holdingHours)
}
