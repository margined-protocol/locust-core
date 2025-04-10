package levana

import (
	"time"

	"github.com/margined-protocol/locust-core/pkg/contracts/levana/factory"
	"github.com/margined-protocol/locust-core/pkg/contracts/levana/market"
)

// MarketData represents a full snapshot of a market
// including config, live status, and any open positions.
type MarketData struct {
	MarketInfo factory.MarketInfoResponse // from factory/types.go
	Status     market.StatusResponse      // from market/types.go
	Positions  *market.PositionsResponse  // from market/types.go
}

// FundingRate represents historical funding rate data
// for long and short positions.
type FundingRate struct {
	Timestamp string `json:"timestamp"`
	LongRate  string `json:"long_rate"`
	ShortRate string `json:"short_rate"`
}

// MarketDecision represents the evaluation result for a specific market and direction.
type MarketDecision struct {
	MarketAddr string
	MarketID   string
	MarketType string
	Direction  string // "short" or "long"

	// Evaluation metrics
	ProfitEstimate    float64
	CurrentRate       float64
	EMARate           float64
	ProjectedPostRate float64
	Fees              float64
	OpenInterest      float64
	Imbalance         float64
	HasPosition       bool
	PositionID        int64

	// Recommended action
	Action    string // "open", "increase", "decrease", "close", or "ignore"
	Note      string
	Timestamp time.Time
}
