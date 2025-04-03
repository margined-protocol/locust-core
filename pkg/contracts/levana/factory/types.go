package factory

type MarketInfoRequest struct {
	MarketID string `json:"market_id"`
}

type MarketInfoResponse struct {
	MarketAddr        string `json:"market_addr"`
	PositionToken     string `json:"position_token"`
	LiquidityTokenLP  string `json:"liquidity_token_lp"`
	LiquidityTokenXLP string `json:"liquidity_token_xlp"`
}

type MarketsRequest struct {
	Limit      *int    `json:"limit,omitempty"`
	StartAfter *string `json:"start_after,omitempty"`
}

type MarketsResponse struct {
	Markets []string `json:"markets"`
}
