package milkyway

type WithdrawMessage struct {
	Withdraw *WithdrawDetails `json:"withdraw"`
}

type WithdrawDetails struct {
	BatchID uint64 `json:"batch_id"`
}

// osmo1f5vfcph2dvfeqcqkhetwv75fda69z7e5c2dldm3kvgj23crkv6wqcn47a0 - milkTIA contract
type BatchResponse struct {
	ID                     uint64 `json:"id"`
	BatchTotalLiquidStake  string `json:"batch_total_liquid_stake"`
	ExpectedNativeUnstaked string `json:"expected_native_unstaked"`
	ReceivedNativeUnstaked string `json:"received_native_unstaked"`
	UnstakeRequestCount    uint64 `json:"unstake_request_count"`
	NextBatchAuctionTime   string `json:"next_batch_action_time"`
	Status                 string `json:"status"`
}

type UnstakeRequestResponse struct {
	Requests []UnstakeRequest
}

type UnstakeRequest struct {
	BatchID uint64 `json:"batch_id"`
	User    string `json:"user"`
	Amount  string `json:"amount"`
}
