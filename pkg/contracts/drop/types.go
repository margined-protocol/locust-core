package drop

type TokensResponse struct {
	Tokens []string `json:"tokens"`
}

type NftInfoResponse struct {
	TokenURI  *string   `json:"token_uri"`
	Extension Extension `json:"extension"`
}

type Extension struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Attributes  []Trait `json:"attributes"`
	BatchID     string  `json:"batch_id"`
	Amount      string  `json:"amount"`
}

type Trait struct {
	DisplayType *string `json:"display_type"`
	TraitType   string  `json:"trait_type"`
	Value       string  `json:"value"`
}

type UnbondBatchResponse struct {
	TotalDassetAmountToWithdraw string           `json:"total_dasset_amount_to_withdraw"`
	ExpectedNativeAssetAmount   string           `json:"expected_native_asset_amount"`
	ExpectedReleaseTime         int64            `json:"expected_release_time"`
	TotalUnbondItems            int              `json:"total_unbond_items"`
	Status                      string           `json:"status"`
	SlashingEffect              *string          `json:"slashing_effect"`
	UnbondedAmount              *string          `json:"unbonded_amount"`
	WithdrawnAmount             *string          `json:"withdrawn_amount"`
	StatusTimestamps            StatusTimestamps `json:"status_timestamps"`
}

// StatusTimestamps represents the status timestamps for different states.
type StatusTimestamps struct {
	New                  int64  `json:"new"`
	UnbondRequested      *int64 `json:"unbond_requested"`
	UnbondFailed         *int64 `json:"unbond_failed"`
	Unbonding            *int64 `json:"unbonding"`
	Withdrawing          *int64 `json:"withdrawing"`
	Withdrawn            *int64 `json:"withdrawn"`
	WithdrawingEmergency *int64 `json:"withdrawing_emergency"`
	WithdrawnEmergency   *int64 `json:"withdrawn_emergency"`
}
