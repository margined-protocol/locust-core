package perps

// ╔═══════════════════════════════════════════════════════════════════════════╗
// ║                               dYdX Types                                  ║
// ╚═══════════════════════════════════════════════════════════════════════════╝

// IndexerSubaccountResponse represents the top-level response from the indexer
type IndexerSubaccountResponse struct {
	Subaccount IndexerSubaccount `json:"subaccount"`
}

// IndexerSubaccount represents a subaccount's details
type IndexerSubaccount struct {
	Address                    string                          `json:"address"`
	SubaccountNumber           int                             `json:"subaccountNumber"`
	Equity                     string                          `json:"equity"`
	FreeCollateral             string                          `json:"freeCollateral"`
	OpenPerpetualPositions     map[string]IndexerPerpPosition  `json:"openPerpetualPositions"`
	AssetPositions             map[string]IndexerAssetPosition `json:"assetPositions"`
	MarginEnabled              bool                            `json:"marginEnabled"`
	UpdatedAtHeight            string                          `json:"updatedAtHeight"`
	LatestProcessedBlockHeight string                          `json:"latestProcessedBlockHeight"`
}

// IndexerPerpPosition represents a perpetual position
type IndexerPerpPosition struct {
	Market           string  `json:"market"`
	Status           string  `json:"status"`
	Side             string  `json:"side"`
	Size             string  `json:"size"`
	MaxSize          string  `json:"maxSize"`
	EntryPrice       string  `json:"entryPrice"`
	ExitPrice        *string `json:"exitPrice"`
	RealizedPnl      string  `json:"realizedPnl"`
	UnrealizedPnl    string  `json:"unrealizedPnl"`
	CreatedAt        string  `json:"createdAt"`
	CreatedAtHeight  string  `json:"createdAtHeight"`
	ClosedAt         *string `json:"closedAt"`
	SumOpen          string  `json:"sumOpen"`
	SumClose         string  `json:"sumClose"`
	NetFunding       string  `json:"netFunding"`
	SubaccountNumber int     `json:"subaccountNumber"`
}

// IndexerAssetPosition represents an asset position
type IndexerAssetPosition struct {
	Size             string `json:"size"`
	Symbol           string `json:"symbol"`
	Side             string `json:"side"`
	AssetId          string `json:"assetId"`
	SubaccountNumber int    `json:"subaccountNumber"`
}

// IndexerCandleResponse represents the top-level response from the indexer
type IndexerCandleResponse struct {
	Candles []IndexerCandle `json:"candles"`
}

// IndexerCandle represents a candle
type IndexerCandle struct {
	StartedAt              string `json:"startedAt"`
	Ticker                 string `json:"ticker"`
	Resolution             string `json:"resolution"`
	Low                    string `json:"low"`
	High                   string `json:"high"`
	Open                   string `json:"open"`
	Close                  string `json:"close"`
	BaseTokenVolume        string `json:"baseTokenVolume"`
	UsdVolume              string `json:"usdVolume"`
	Trades                 int    `json:"trades"`
	StartingOpenInterest   string `json:"startingOpenInterest"`
	OrderbookMidPriceOpen  string `json:"orderbookMidPriceOpen"`
	OrderbookMidPriceClose string `json:"orderbookMidPriceClose"`
}

// IndexerFillResponse represents a response containing fills from the dYdX indexer
type IndexerFillResponse struct {
	Fills []IndexerFill `json:"fills"`
}

// IndexerFill represents a single fill from the dYdX indexer
type IndexerFill struct {
	ID                string `json:"id"`
	Side              string `json:"side"`
	Liquidity         string `json:"liquidity"`
	Type              string `json:"type"`
	Market            string `json:"market"`
	MarketType        string `json:"marketType"`
	Price             string `json:"price"`
	Size              string `json:"size"`
	Fee               string `json:"fee"`
	AffiliateRevShare string `json:"affiliateRevShare"`
	CreatedAt         string `json:"createdAt"`
	CreatedAtHeight   string `json:"createdAtHeight"`
	OrderID           string `json:"orderId"`
	ClientMetadata    string `json:"clientMetadata"`
	SubaccountNumber  uint32 `json:"subaccountNumber"`
}
