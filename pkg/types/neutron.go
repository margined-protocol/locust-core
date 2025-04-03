package types

type NeutronGridStrategyConfig struct {
	Token0               string          `toml:"token0" mapstructure:"token0"` // Token0 is the first token in the pair
	Token1               string          `toml:"token1" mapstructure:"token1"` // Token1 is the second token in the pair
	ChainID              string          `toml:"chain_id" mapstructure:"chain_id"`
	Levels               int             `toml:"levels" mapstructure:"levels"`
	LowerBound           float64         `toml:"lower_bound" mapstructure:"lower_bound"`
	UpperBound           float64         `toml:"upper_bound" mapstructure:"upper_bound"`
	DepositFactor        float64         `toml:"deposit_factor" mapstructure:"deposit_factor"`
	InvertBase           bool            `toml:"invert_base" mapstructure:"invert_base"`
	CoinGecko            CoinGeckoConfig `toml:"coingecko" mapstructure:"coingecko"` // CoinGecko configuration
	Frequency            int             `toml:"frequency" mapstructure:"frequency"`
	Granter              string          `toml:"granter" mapstructure:"granter"`
	CancelAllLimitOrders bool            `toml:"cancel_all_limit_orders" mapstructure:"cancel_all_limit_orders"`
}

func (p *NeutronGridStrategyConfig) GetPairID() string {
	return p.Token0 + "<>" + p.Token1
}

type NeutronDexStrategyConfig struct {
	Granter          string        `toml:"granter" mapstructure:"granter"`
	Frequency        int           `toml:"frequency" mapstructure:"frequency"`
	FundContract     string        `toml:"fund_contract" mapstructure:"fund_contract"`
	StrategyContract string        `toml:"strategy_contract" mapstructure:"strategy_contract"`
	Binance          BinanceConfig `toml:"binance" mapstructure:"binance"`
	Pair             PairConfig    `toml:"pair" mapstructure:"pair"`
	WithdrawAll      bool          `toml:"withdraw_all" mapstructure:"withdraw_all"`
	DepositFactor    float64       `toml:"deposit_factor" mapstructure:"deposit_factor"`
	Fees             []uint64      `toml:"fees" mapstructure:"fees"`
	AdjustmentFactor float64       `toml:"adjustment_factor" mapstructure:"adjustment_factor"`
}

type NeutronXemmStrategyConfig struct {
	Granter     string `toml:"granter" mapstructure:"granter"`
	OrderExpiry int64  `toml:"order_expiry" mapstructure:"order_expiry"`

	Token0         string        `toml:"token0" mapstructure:"token0"`                     // Token0 is the first token in the pair
	Token1         string        `toml:"token1" mapstructure:"token1"`                     // Token1 is the second token in the pair
	Spread         int64         `toml:"spread" mapstructure:"spread"`                     // The spread BPS the we target with balanced inventory
	OrderWidth     int64         `toml:"order_width" mapstructure:"order_width"`           // The increment BPS we increment prices by
	OrderNumber    int64         `toml:"order_number" mapstructure:"order_number"`         // The number of orders we create on each side
	MinTradeAmount int64         `toml:"min_trade_amount" mapstructure:"min_trade_amount"` // The smallest amount of the token we will trade
	Sensitivity    string        `toml:"sensitivity" mapstructure:"sensitivity"`           // Sensitivity to the change in price
	Binance        BinanceConfig `toml:"binance" mapstructure:"binance"`
	ActiveTaker    bool          `toml:"active_taker" mapstructure:"active_taker"` // if active is true then we will make trades on taker market
}

// Method to get the pair ID
func (p *NeutronXemmStrategyConfig) GetPairID() string {
	return p.Token0 + "<>" + p.Token1
}

type LimitMmStrategyConfig struct {
	Granter          string        `toml:"granter" mapstructure:"granter"`
	Frequency        int           `toml:"frequency" mapstructure:"frequency"`
	OrderExpiry      int64         `toml:"order_expiry" mapstructure:"order_expiry"`
	Token0           string        `toml:"token0" mapstructure:"token0"` // Token0 is the first token in the pair
	Token0Exponent   int           `toml:"token0" mapstructure:"token0_exponent"`
	Token1           string        `toml:"token1" mapstructure:"token1"` // Token1 is the second token in the pair
	Token1Exponent   int           `toml:"token0" mapstructure:"token1_exponent"`
	Spread           int64         `toml:"spread" mapstructure:"spread"`                     // The spread BPS the we target with balanced inventory
	OrderWidth       int64         `toml:"order_width" mapstructure:"order_width"`           // The increment BPS we increment prices by
	OrderNumber      int64         `toml:"order_number" mapstructure:"order_number"`         // The number of orders we create on each side
	MinTradeAmount   int64         `toml:"min_trade_amount" mapstructure:"min_trade_amount"` // The smallest amount of the token we will trade
	Sensitivity      string        `toml:"sensitivity" mapstructure:"sensitivity"`           // Sensitivity to the change in price
	Binance          BinanceConfig `toml:"binance" mapstructure:"binance"`
	Drop             DropConfig    `toml:"drop" mapstructure:"drop"`
	FundContract     string        `toml:"fund_contract" mapstructure:"fund_contract"`
	StrategyContract string        `toml:"strategy_contract" mapstructure:"strategy_contract"`
}

// Method to get the pair ID
func (p *LimitMmStrategyConfig) GetPairID() string {
	return p.Token0 + "<>" + p.Token1
}

type DropRedemptionStrategyConfig struct {
	Granter          string          `toml:"granter" mapstructure:"granter"`
	FundContract     string          `toml:"fund_contract" mapstructure:"fund_contract"`
	StrategyContract string          `toml:"strategy_contract" mapstructure:"strategy_contract"`
	Frequency        int             `toml:"frequency" mapstructure:"frequency"`
	Jitter           int             `toml:"jitter" mapstructure:"jitter"`
	Astroport        AstroportConfig `toml:"astroport" mapstructure:"astroport"`
	Drop             DropConfig      `toml:"drop" mapstructure:"drop"`
	DryRun           bool            `toml:"dry_run" mapstructure:"dry_run"`
	PropTrade        bool            `toml:"prop_trade" mapstructure:"prop_trade"`
	ArbThreshold     float64         `toml:"arb_threshold" mapstructure:"arb_threshold"`
	MinBuySize       SdkInt          `toml:"min_buy_size" mapstructure:"min_buy_size"`
}

type DexArbitrageStrategyConfig struct {
	Granter       string          `toml:"granter" mapstructure:"granter"`
	Frequency     int             `toml:"frequency" mapstructure:"frequency"`
	MinTradeSize  SdkInt          `toml:"min_trade_size" mapstructure:"min_trade_size"`
	Slippage      float64         `toml:"slippage" mapstructure:"slippage"`
	ArbThreshold  float64         `toml:"arb_threshold" mapstructure:"arb_threshold"`
	AstroportBuy  AstroportConfig `toml:"astroport_buy" mapstructure:"astroport_buy"`
	AstroportSell AstroportConfig `toml:"astroport_sell" mapstructure:"astroport_sell"`
	DexBuy        DexConfig       `toml:"dex_buy" mapstructure:"dex_buy"`
	DexSell       DexConfig       `toml:"dex_sell" mapstructure:"dex_sell"`
}

type CoinHedgeStrategyConfig struct {
	Kill              bool         `toml:"kill" mapstructure:"kill"`                   // will kill the strategy, not for production use
	FundContract      string       `toml:"fund_contract" mapstructure:"fund_contract"` // Fund contract address, if not prop trading
	ChainID           string       `toml:"chain_id" mapstructure:"chain_id"`
	Name              string       `toml:"name" mapstructure:"name"` // some human intelligible name for the strategy, e.g. "ATOM Hedge Hydro"
	ExecutionInterval int          `toml:"execution_interval" mapstructure:"execution_interval"`
	PerpDex           PerpDex      `toml:"perp_dex" mapstructure:"perp_dex"`
	Hedge             HedgeConfig  `toml:"hedge_config" mapstructure:"hedge_config"`
	Skip              SkipConfig   `toml:"skip_config" mapstructure:"skip_config"`
	Slinky            SlinkyConfig `toml:"slinky_config" mapstructure:"slinky_config"`
	DB                DBConfig     `toml:"db_config" mapstructure:"db_config"`

	// Vault contracts
	Vault LocustVault `toml:"vault" mapstructure:"vault"`

	// Eventually I think we should just have a single chain config and then a list of chains
	Neutron Chain `toml:"neutron" mapstructure:"neutron"`
	Dydx    Chain `toml:"dydx" mapstructure:"dydx"`
	Noble   Chain `toml:"noble" mapstructure:"noble"`
}
