package types

type LendLeaseStrategyConfig struct {
	ChainID string `toml:"chain_id" mapstructure:"chain_id"`
	Kill    bool   `toml:"kill" mapstructure:"kill"`
	// Eventually I think we should just have a single chain config and then a list of chains
	Osmosis     Chain      `toml:"osmosis" mapstructure:"osmosis"`
	Neutron     Chain      `toml:"neutron" mapstructure:"neutron"`
	Noble       Chain      `toml:"noble" mapstructure:"noble"`
	Umee        Chain      `toml:"umee" mapstructure:"umee"`
	MarsOsmosis MarsConfig `toml:"mars_osmosis" mapstructure:"mars_osmosis"`
	MarsNeutron MarsConfig `toml:"mars_neutron" mapstructure:"mars_neutron"`
	UmeeMarket  UmeeConfig `toml:"umee_market" mapstructure:"umee_market"`

	// Vault contracts
	Vault LocustVault `toml:"vault" mapstructure:"vault"`

	MinRateDeltaBPS       uint64 `toml:"min_rate_delta_bps" mapstructure:"min_rate_delta_bps"`
	MaxRateImpactBPS      uint64 `toml:"max_rate_impact_bps" mapstructure:"max_rate_impact_bps"`
	MinRebalanceAmount    uint64 `toml:"min_rebalance_amount" mapstructure:"min_rebalance_amount"`
	MaxRebalanceAmount    uint64 `toml:"max_rebalance_amount" mapstructure:"max_rebalance_amount"`
	MaxRepaymentAmount    uint64 `toml:"max_repayment_amount" mapstructure:"max_repayment_amount"`
	MinDistributionAmount uint64 `toml:"min_distribution_amount" mapstructure:"min_distribution_amount"`

	// Cycle Params
	RebalanceInterval uint64 `toml:"rebalance_interval" mapstructure:"rebalance_interval"`
	RepaymentInterval uint64 `toml:"repayment_interval" mapstructure:"repayment_interval"`
	SweepInterval     uint64 `toml:"sweep_interval" mapstructure:"sweep_interval"`
}

type CLMMStrategyConfig struct {
	DefaultToken0Amount SdkInt `toml:"default_token_0_amount" mapstructure:"default_token_0_amount"`
	DefaultToken1Amount SdkInt `toml:"default_token_1_amount" mapstructure:"default_token_1_amount"`
	ChainID             string `toml:"chain_id" mapstructure:"chain_id"`
	Kill                bool   `toml:"kill" mapstructure:"kill"`
	Granter             string `toml:"granter" mapstructure:"granter"`
	Pool                Pool   `toml:"pool" mapstructure:"pool"`

	// NOTE: for the steak strategies (Eris, Backbone) the pricing_url is a contract
	PricingURL       string        `toml:"pricing_url" mapstructure:"pricing_url"`
	PriceScaleFactor int64         `toml:"price_scale_factor" mapstructure:"price_scale_factor"`
	Binance          BinanceConfig `toml:"binance" mapstructure:"binance"`
	Pryzm            PryzmConfig   `toml:"pryzm" mapstructure:"pryzm"`
	Drop             Drop          `toml:"drop" mapstructure:"drop"`
	APRURL           string        `toml:"apr_url" mapstructure:"apr_url"`
	Position         Position      `toml:"position" mapstructure:"position"`
	IsLSD            bool          `toml:"is_lsd" mapstructure:"is_lsd"`
}

type GridStrategyConfig struct {
	DefaultToken0Amount SdkInt `toml:"default_token_0_amount" mapstructure:"default_token_0_amount"`
	DefaultToken1Amount SdkInt `toml:"default_token_1_amount" mapstructure:"default_token_1_amount"`
	ChainID             string `toml:"chain_id" mapstructure:"chain_id"`
	Granter             string `toml:"granter" mapstructure:"granter"`
	Pool                Pool   `toml:"pool" mapstructure:"pool"`
	Grid                Grid   `toml:"grid" mapstructure:"grid"`
}

type RedemptionStrategyConfig struct {
	LSPName        string          `toml:"lsp_name" mapstructure:"lsp_name"`
	Buffer         SdkInt          `toml:"buffer" mapstructure:"buffer"`
	Discount       int64           `toml:"discount" mapstructure:"discount"`   // Discount we try to trade to (BPS)
	Precision      int64           `toml:"precision" mapstructure:"precision"` // Decimal places of base assets
	Interval       int64           `toml:"interval" mapstructure:"interval"`   // Minutes between automated running of strategy
	MaxSwapAmount  SdkInt          `toml:"max_swap_amount" mapstructure:"max_swap_amount"`
	Threshold      int64           `toml:"threshold" mapstructure:"threshold"` // BPS
	ChainID        string          `toml:"chain_id" mapstructure:"chain_id"`
	Vault          string          `toml:"vault" mapstructure:"vault"`
	Float          SdkInt          `toml:"float" mapstructure:"float"`
	Pool           Pool            `toml:"pool" mapstructure:"pool"`
	PricingURL     string          `toml:"pricing_url" mapstructure:"pricing_url"`
	FromBase       IBCTransfer     `toml:"from_base" mapstructure:"from_base"`
	ToLSP          IBCTransfer     `toml:"to_lsp" mapstructure:"to_lsp"`
	UnbondStride   *UnbondStride   `toml:"unbond_stride" mapstructure:"unbond_stride"`
	UnbondDrop     *UnbondDrop     `toml:"unbond_drop" mapstructure:"unbond_drop"`
	UnbondMilkyway *UnbondMilkyway `toml:"unbond_milkyway" mapstructure:"unbond_milkyway"`
	LSD            Chain           `toml:"lsd_chain" mapstructure:"lsd_chain"`
	Base           Chain           `toml:"base_chain" mapstructure:"base_chain"`
}

type IBCTransfer struct {
	Port            string `toml:"port" mapstructure:"port"`
	Channel         string `toml:"channel" mapstructure:"channel"`
	Denom           string `toml:"denom" mapstructure:"denom"`
	Buffer          SdkInt `toml:"buffer" mapstructure:"buffer"`
	Timeoutheight   uint64 `toml:"timeout_height" mapstructure:"timeout_height"`
	Timeoutduration uint64 `toml:"timeout_duration" mapstructure:"timeout_duration"`
	Memo            string `toml:"memo" mapstructure:"memo"`
}

type UnbondStride struct {
	Denom    string `toml:"denom" mapstructure:"denom"`
	Buffer   SdkInt `toml:"buffer" mapstructure:"buffer"`
	HostZone string `toml:"host_zone" mapstructure:"host_zone"`
}

type UnbondDrop struct {
	Denom                            string `toml:"denom" mapstructure:"denom"`
	CoreContractAddress              string `json:"core_contract_address" mapstructure:"core_contract_address"`
	WithdrawalVoucherContractAddress string `json:"withdrawal_voucher_contract_address" mapstructure:"withdrawal_voucher_contract_address"`
	WithdrawalManagerContractAddress string `json:"withdrawal_manager_contract_address" mapstructure:"withdrawal_manager_contract_address"`
}

type UnbondMilkyway struct {
	Contract string `toml:"contract" mapstructure:"contract"`
	Denom    string `toml:"denom" mapstructure:"denom"`
}
