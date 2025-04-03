package types

import (
	"github.com/margined-protocol/locust-core/pkg/db"
)

type ZoneResponse struct {
	Zone *Zone `json:"zone"`
}

type PairConfig struct {
	Token0          string  `toml:"token0" mapstructure:"token0"`
	Token1          string  `toml:"token0" mapstructure:"token1"`
	InvertBase      bool    `toml:"invert_base" mapstructure:"invert_base"`
	Price           string  `toml:"price" mapstructure:"price"`
	Ticks           int     `toml:"ticks" mapstructure:"ticks"`
	Fees            []int   `toml:"fees" mapstructure:"fees"`
	RebalanceFactor float64 `toml:"rebalance_factor" mapstructure:"rebalance_factor"`
	DepositFactor   float64 `toml:"deposit_factor" mapstructure:"deposit_factor"`
	SwapFactor      float64 `toml:"swap_factor" mapstructure:"swap_factor"`
	SwapAccuracy    float64 `toml:"swap_accuracy" mapstructure:"swap_accuracy"`
	DepositAccuracy float64 `toml:"deposit_accuracy" mapstructure:"deposit_accuracy"`
	Amplitude1      float64 `toml:"amplitude1" mapstructure:"amplitude1"`
	Period1         float64 `toml:"period1" mapstructure:"period1"`
	Amplitude2      float64 `toml:"amplitude2" mapstructure:"amplitude2"`
	Period2         float64 `toml:"period2" mapstructure:"period2"`
}

// Method to get the pair ID
func (p *PairConfig) GetPairID() string {
	return p.Token0 + "<>" + p.Token1
}

type CoinGeckoConfig struct {
	URL      string `toml:"coingecko_url" mapstructure:"coingecko_url"`         // URL for the CoinGecko API
	APIKey   string `toml:"coingecko_api_key" mapstructure:"coingecko_api_key"` // Coingecko API key
	TokenID0 string `toml:"token_id_0" mapstructure:"token_id_0"`               // Base token
	TokenID1 string `toml:"token_id_1" mapstructure:"token_id_1"`               // Quote token
	Currency string `toml:"currency" mapstructure:"currency"`                   // Currency to show the price in
}

type BinanceConfig struct {
	APIKey    string `toml:"binance_api_key" mapstructure:"binance_api_key"`
	SecretKey string `toml:"binance_secret_key" mapstructure:"binance_secret_key"`
	Symbol    string `toml:"symbol" mapstructure:"symbol"`
	Token0    string `toml:"token0" mapstructure:"token0"` // Token0 is the base token in the pair
	Token1    string `toml:"token1" mapstructure:"token1"` // Token1 is the quote token in the pair
}

type SlinkyConfig struct {
	Base  string `toml:"base" mapstructure:"base"`   // base token in the pair
	Quote string `toml:"quote" mapstructure:"quote"` // quote token in the pair
}

type Drop struct {
	Chain Chain      `toml:"chain" mapstructure:"chain"`
	Info  UnbondDrop `toml:"info" mapstructure:"info"`
}

type PryzmConfig struct {
	BaseURL      string `toml:"base_url" mapstructure:"base_url"`
	PoolID       string `toml:"pool_id" mapstructure:"pool_id"`
	StakingDenom string `toml:"staking_denom" mapstructure:"staking_denom"`
	LPDenom      string `toml:"lp_denom" mapstructure:"lp_denom"`
	BaseDenom    string `toml:"base_denom" mapstructure:"base_denom"`
}

type Zone struct {
	RedemptionRate     string `json:"redemption_rate"`
	LastRedemptionRate string `json:"last_redemption_rate"`
}

type Grid struct {
	Levels     int     `toml:"levels" mapstructure:"levels"`
	LowerBound float64 `toml:"lower_bound" mapstructure:"lower_bound"`
	UpperBound float64 `toml:"upper_bound" mapstructure:"upper_bound"`
}

type Pool struct {
	ID         uint64 `toml:"id" mapstructure:"id"`
	BaseDenom  string `toml:"base_denom" mapstructure:"base_denom"`
	QuoteDenom string `toml:"quote_denom" mapstructure:"quote_denom"`
}

type Position struct {
	AskSpread           int64     `toml:"ask_spread" mapstructure:"ask_spread"` // The spread BPS for asks
	AskLiquidityProfile []float64 `toml:"ask_liquidity_profile" mapstructure:"ask_liquidity_profile"`
	BidSpread           int64     `toml:"bid_spread" mapstructure:"bid_spread"` // The spread BPS for bids
	BidLiquidityProfile []float64 `toml:"bid_liquidity_profile" mapstructure:"bid_liquidity_profile"`
	Discount            int64     `toml:"discount" mapstructure:"discount"`       // Discount applied to external price BPS
	MaxSpread           int64     `toml:"max_spread" mapstructure:"max_spread"`   // Max spread BPS allowed between best prices
	Sensitivity         int64     `toml:"sensitivity" mapstructure:"sensitivity"` // Sensitivity to price changes BPS
}

type AstroportConfig struct {
	ContractAddress string `json:"contract_address" mapstructure:"contract_address"`
	OfferAsset      string `json:"offer_asset" mapstructure:"offer_asset"`
	AskAsset        string `json:"ask_asset" mapstructure:"ask_asset"`
	Exponent        int    `json:"exponent" mapstructure:"exponent"`
	PoolType        string `json:"pool_type mapstructure:pooltype"`
}

type DropConfig struct {
	CoreContractAddress              string `json:"core_contract_address" mapstructure:"core_contract_address"`
	WithdrawalVoucherContractAddress string `json:"withdrawal_voucher_contract_address" mapstructure:"withdrawal_voucher_contract_address"`
	WithdrawalManagerContractAddress string `json:"withdrawal_manager_contract_address" mapstructure:"withdrawal_manager_contract_address"`
}

type DexConfig struct {
	TokenIn  string `json:"token_in" mapstructure:"token_in"`
	TokenOut string `json:"token_out" mapstructure:"token_out"`
}

func (p *DexConfig) GetPairID() string {
	return p.TokenIn + "<>" + p.TokenOut
}

type LevanaCCStrategyConfig struct {
	DB                db.Config `toml:"db" mapsturcture:"db"`
	Granter           string    `toml:"granter" mapstructure:"granter"`
	Frequency         int       `toml:"frequency" mapstructure:"frequency"`
	FactoryContract   string    `toml:"factory_contract" mapstructure:"factory_contract"`
	LevanaAPIURL      string    `toml:"levana_api_url" mapstructure:"levana_api_url"`
	BaseDenom         string    `toml:"base_denom" mapstructure:"base_denom"`
	BaseDenomExponent int       `toml:"base_denom_exponent" mapstructure:"base_denom_exponent"`
	PoolID            uint64    `toml:"pool_id" mapstructure:"pool_id"`
	TradeSize         int64     `toml:"trade_size" mapstructure:"trade_size"`
	Leverage          int64     `toml:"leverage" mapstructure:"leverage"`
	Shutdown          bool      `toml:"shutdown" mapstructure:"shutdown"`
	Slippage          string    `toml:"slippage" mapstructure:"slippage"`
	MaxPriceDeviation string    `toml:"max_price_deviation" mapstructure:"max_price_deviation"`
	MinFundingRate    string    `toml:"min_funding_rate" mapstructure:"min_funding_rate"`
	MinOpenInterest   string    `toml:"min_open_interest" mapstructure:"min_open_interest"`
	MaxPositions      int       `toml:"max_positions" mapstructure:"max_positions"`
	TargetMarkets     []string  `toml:"target_markets" mapstructure:"target_markets"`
	FundContract      string    `toml:"fund_contract" mapstructure:"fund_contract"`
	StrategyContract  string    `toml:"strategy_contract" mapstructure:"strategy_contract"`
}

type SkipConfig struct {
	URL            string `json:"url" mapstructure:"url"`                           // url of skip api
	SwapEntryPoint string `json:"swap_entry_point" mapstructure:"swap_entry_point"` // swap entry point contract for the chain
}

type DBConfig struct {
	Dsn string `json:"dsn" mapstructure:"dsn"` // database connection details
}

type HedgeConfig struct {
	TargetAmount      SdkInt  `json:"target_amount" mapstructure:"target_amount"`             // total amount to be hedged
	MaxTradeAmount    SdkInt  `json:"max_trade_amount" mapstructure:"max_trade_amount"`       // maximum amount to be traded in a single trade
	MaxPriceImpact    float64 `json:"max_price_impact" mapstructure:"max_price_impact"`       // maximum slippage allowed, 1 = 1%, 0.5 = 0.5%
	TargetMarginRatio float64 `json:"target_margin_ratio" mapstructure:"target_margin_ratio"` // target margin ratio
	TokenIn           string  `json:"token_in" mapstructure:"token_in"`                       // token to be hedged
	TokenOut          string  `json:"token_out" mapstructure:"token_out"`                     // token to hedge against
	TokenInDecimals   int     `json:"token_in_decimals" mapstructure:"token_in_decimals"`     // decimals of token_in
	TokenOutDecimals  int     `json:"token_out_decimals" mapstructure:"token_out_decimals"`   // decimals of token_out
	MarginThreshold   int64   `json:"margin_threshold" mapstructure:"margin_threshold"`       // BPS proximity to liquidation price we accept prior to adding more margin
	Expiration        string  `json:"expiration" mapstructure:"expiration"`                   // expiration of hedge as date `YYYY-MM-DD HH:MM:SS`
}

type MarsConfig struct {
	CreditAccount uint64 `json:"credit_account" mapstructure:"credit_account"`
	CreditManager string `json:"credit_manager" mapstructure:"credit_manager"`
	RedBank       string `json:"red_bank" mapstructure:"red_bank"`
	Oracle        string `json:"oracle" mapstructure:"oracle"`
	Perps         string `json:"perps" mapstructure:"perps"`
	OracleDenom   string `json:"oracle_denom" mapstructure:"oracle_denom"` // Denom of the oracle price, e.g. untrn or ibc/xyz...
	Market        string `json:"market" mapstructure:"market"`
}

type DydxConfig struct {
	Market               string `json:"market" mapstructure:"market"`
	IndexerURL           string `json:"indexer_url" mapstructure:"indexer_url"`
	MarketID             uint32 `json:"market_id" mapstructure:"market_id"`
	SubaccountID         uint32 `json:"subaccount_id" mapstructure:"subaccount_id"`
	QuantumConversionExp int64  `json:"quantum_conversion_exp" mapstructure:"quantum_conversion_exp"`
	SubticksPerTick      uint64 `json:"subticks_per_tick" mapstructure:"subticks_per_tick"`
	StepBaseQuantums     uint64 `json:"step_base" mapstructure:"step_base_quantums"`
	AtomicResolution     int64  `json:"atomic_resolution" mapstructure:"atomic_resolution"`
	Denom                string `json:"denom" mapstructure:"denom"`
	MinEquity            SdkInt `json:"min_equity" mapstructure:"min_equity"` // minimum equity required to maintain a position
}

type PerpDex struct {
	Provider   string      `json:"provider" mapstructure:"provider"`
	MarsConfig *MarsConfig `json:"mars_config" mapstructure:"mars_config"`
	DydxConfig *DydxConfig `json:"dydx_config" mapstructure:"dydx_config"`
}
