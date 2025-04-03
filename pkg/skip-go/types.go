package skipgo

import (
	"context"
	"fmt"
	"math/big"

	"go.uber.org/zap"
)

// TransactionState represents the state of a transaction.
type TransactionState string

const (
	StateSubmitted        TransactionState = "STATE_SUBMITTED"
	StatePending          TransactionState = "STATE_PENDING"
	StateCompletedSuccess TransactionState = "STATE_COMPLETED_SUCCESS"
	StateCompletedError   TransactionState = "STATE_COMPLETED_ERROR"
	StateAbandoned        TransactionState = "STATE_ABANDONED"
	StatePendingError     TransactionState = "STATE_PENDING_ERROR"
)

// IsCompleted checks if the transaction is completed.
func (s TransactionState) IsCompleted() bool {
	return s == StateCompletedSuccess || s == StateCompletedError || s == StateAbandoned || s == StatePendingError
}

// IsCompletedError checks if the transaction completed with an error.
func (s TransactionState) IsCompletedError() bool {
	return s == StateCompletedError || s == StateAbandoned || s == StatePendingError
}

// TxHash represents a transaction hash.
type TxHash string

// SkipGoClient defines the interface for SkipGo operations.
type Client interface {
	Balance(ctx context.Context, request *BalancesRequest) (*BalancesResponse, error)
	Route(ctx context.Context, sourceAssetDenom, sourceAssetChainID, destAssetDenom, destAssetChainID string, amountIn *big.Int) (*RouteResponse, error)
	SwapRoute(ctx context.Context, tokenIn, tokenOut, chainID string, amountIn *big.Int) (*RouteResponse, error)
	FindOptimalSwapRoute(
		ctx context.Context, logger *zap.Logger, tokenIn, tokenOut, chainID string, amountIn *big.Int, maxPriceImpact float64,
	) (*big.Int, *RouteResponse, error)
	Msgs(ctx context.Context, route RouteResponse, addressList []string, slippage string) ([]Tx, error)
	SubmitTx(ctx context.Context, tx []byte, chainID string) (TxHash, error)
	TrackTx(ctx context.Context, txHash, chainID string) (TxHash, error)
	Status(ctx context.Context, tx TxHash, chainID string) (*StatusResponse, error)
}

// BalancesRequest represents a request for balances.
type BalancesRequest struct {
	Chains map[string]ChainRequest `json:"chains"`
}

// ChainRequest represents a single chain's balance request.
type ChainRequest struct {
	Address string   `json:"address"`
	Denoms  []string `json:"denoms"`
}

// BalancesResponse represents the response for a balance request.
type BalancesResponse struct {
	Chains map[string]ChainResponse `json:"chains"`
}

// ChainResponse represents a single chain's balance response.
type ChainResponse struct {
	Address string                 `json:"address"`
	Denoms  map[string]DenomDetail `json:"denoms"`
}

// DenomDetail provides details about a specific denomination.
type DenomDetail struct {
	Amount          string `json:"amount"`
	Decimals        uint8  `json:"decimals"`
	FormattedAmount string `json:"formatted_amount"`
	Price           string `json:"price"`
	ValueUSD        string `json:"value_usd"`
}

type SmartSwapOptions struct {
	EVMSwaps    bool `json:"evm_swaps"`
	SplitRoutes bool `json:"split_routes"`
}

// RouteResponse represents the the request for a route.
type RouteRequest struct {
	SourceAssetDenom   string           `json:"source_asset_denom"`
	SourceAssetChainID string           `json:"source_asset_chain_id"`
	DestAssetDenom     string           `json:"dest_asset_denom"`
	DestAssetChainID   string           `json:"dest_asset_chain_id"`
	AmountIn           string           `json:"amount_in"`
	AllowMultiTx       bool             `json:"allow_multi_tx"`
	Bridges            []string         `json:"bridges"`
	AllowUnsafe        bool             `json:"allow_unsafe"`
	SmartSwapOptions   SmartSwapOptions `json:"smart_swap_options"`
}

// RouteResponse represents the response from a route request.
type RouteResponse struct {
	AmountIn               string   `json:"amount_in"`
	AmountOut              string   `json:"amount_out"`
	SourceAssetDenom       string   `json:"source_asset_denom"`
	SourceAssetChainID     string   `json:"source_asset_chain_id"`
	DestAssetDenom         string   `json:"dest_asset_denom"`
	DestAssetChainID       string   `json:"dest_asset_chain_id"`
	Operations             []any    `json:"operations"`
	ChainIDs               []string `json:"chain_ids"`
	RequiredChainAddresses []string `json:"required_chain_addresses"`
	DoesSwap               bool     `json:"does_swap"`
	EstimatedAmountOut     string   `json:"estimated_amount_out"`
	TxsRequired            int      `json:"txs_required"`
	USDAmountIn            string   `json:"usd_amount_in"`
	USDAmountOut           string   `json:"usd_amount_out"`
	SwapPriceImpactPercent string   `json:"swap_price_impact_percent"`
}

// EVMTx represents an Ethereum transaction.
type EVMTx struct {
	ChainID                string          `json:"chain_id"`
	To                     string          `json:"to"`
	Value                  string          `json:"value"`
	Data                   string          `json:"data"`
	SignerAddress          string          `json:"signer_address"`
	RequiredERC20Approvals []ERC20Approval `json:"required_erc20_approvals"`
}

// ERC20Approval represents an ERC20 token approval.
type ERC20Approval struct {
	TokenContract string `json:"token_contract"`
	Spender       string `json:"spender"`
	Amount        string `json:"amount"`
}

// CosmosMessage represents a Cosmos SDK message.
type CosmosMessage struct {
	Msg        string `json:"msg"`
	MsgTypeURL string `json:"msg_type_url"`
}

// CosmosTx represents a Cosmos SDK transaction.
type CosmosTx struct {
	ChainID       string          `json:"chain_id"`
	Path          []string        `json:"path"`
	SignerAddress string          `json:"signer_address"`
	Msgs          []CosmosMessage `json:"msgs"`
}

// Tx represents a generic transaction that can be either EVM or Cosmos.
type Tx struct {
	EVMTx             *EVMTx    `json:"evm_tx,omitempty"`
	CosmosTx          *CosmosTx `json:"cosmos_tx,omitempty"`
	OperationsIndices []int     `json:"operations_indices"`
}

// MsgsRequest represents a request to fetch messages for a transaction.
type MsgsRequest struct {
	SourceAssetDenom         string   `json:"source_asset_denom"`
	SourceAssetChainID       string   `json:"source_asset_chain_id"`
	DestAssetDenom           string   `json:"dest_asset_denom"`
	DestAssetChainID         string   `json:"dest_asset_chain_id"`
	AmountIn                 string   `json:"amount_in"`
	AmountOut                string   `json:"amount_out"`
	SlippageTolerancePercent string   `json:"slippage_tolerance_percent"`
	AddressList              []string `json:"address_list"`
	Operations               []any    `json:"operations"`
}

// MsgsResponse represents the response from a Msgs request.
type MsgsResponse struct {
	Txs []Tx `json:"txs"`
}

// SubmitRequest represents a request to submit a transaction.
type SubmitRequest struct {
	Tx      string `json:"tx"`
	ChainID string `json:"chain_id"`
}

// SubmitResponse represents the response after submitting a transaction.
type SubmitResponse struct {
	TxHash string `json:"tx_hash"`
}

// TrackRequest represents a request to track a transaction.
type TrackRequest struct {
	TxHash  string `json:"tx_hash"`
	ChainID string `json:"chain_id"`
}

// TrackResponse represents the response after tracking a transaction.
type TrackResponse struct {
	TxHash string `json:"tx_hash"`
}

// StatusResponse represents the status of a transaction.
type StatusResponse struct {
	Transfers []Transfer `json:"transfers"`
}

// Transfer represents a single transfer within a transaction.
type Transfer struct {
	State                TransactionState     `json:"state"`
	TransferSequence     []TransferSequence   `json:"transfer_sequence"`
	NextBlockingTransfer *TransferSequence    `json:"next_blocking_transfer,omitempty"`
	TransferAssetRelease TransferAssetRelease `json:"transfer_asset_release,omitempty"`
	Error                *string              `json:"error,omitempty"`
}

// TransferSequence represents the sequence of transfers.
type TransferSequence struct {
	IBCTransfer       *IBCTransfer     `json:"ibc_transfer,omitempty"`
	AxelarTransfer    *AxelarTransfer  `json:"axelar_transfer,omitempty"`
	CCTPTransfer      *GenericTransfer `json:"cctp_transfer,omitempty"`
	HyperlaneTransfer *GenericTransfer `json:"hyperlane_transfer,omitempty"`
	OpinitTransfer    *GenericTransfer `json:"opinit_transfer,omitempty"`
}

// TransferAssetRelease represents the release of transfer assets.
type TransferAssetRelease struct {
	ChainID  string `json:"chain_id"`
	Denom    string `json:"denom"`
	Released bool   `json:"released"`
}

// StatusTrackingTx represents a transaction being tracked.
type StatusTrackingTx struct {
	ChainID      string `json:"chain_id"`
	ExplorerLink string `json:"explorer_link"`
	TxHash       string `json:"tx_hash"`
}

// IBCTransfer represents an IBC transfer.
type IBCTransfer struct {
	FromChainID string    `json:"from_chain_id"`
	ToChainID   string    `json:"to_chain_id"`
	State       string    `json:"state"`
	PacketTxs   PacketTxs `json:"packet_txs"`
}

// PacketTxs represents the transactions involved in a packet.
type PacketTxs struct {
	SendTx        *StatusTrackingTx `json:"send_tx"`
	ReceiveTx     *StatusTrackingTx `json:"receive_tx"`
	AcknowledgeTx *StatusTrackingTx `json:"acknowledge_tx"`
	TimeoutTx     *StatusTrackingTx `json:"timeout_tx,omitempty"`
	Error         *string           `json:"error,omitempty"`
}

// AxelarTransfer represents an Axelar transfer.
type AxelarTransfer struct {
	FromChainID    string       `json:"from_chain_id"`
	ToChainID      string       `json:"to_chain_id"`
	Type           string       `json:"type"`
	State          string       `json:"state"`
	Txs            SendTokenTxs `json:"txs"`
	AxelarScanLink string       `json:"axelar_scan_link"`
}

// SendTokenTxs represents the token transactions in an Axelar transfer.
type SendTokenTxs struct {
	SendTx    *StatusTrackingTx `json:"send_tx"`
	ConfirmTx *StatusTrackingTx `json:"confirm_tx,omitempty"`
	ExecuteTx *StatusTrackingTx `json:"execute_tx"`
	Error     *string           `json:"error,omitempty"`
}

// GenericTransfer represents a generic transfer.
type GenericTransfer struct {
	ToChain   string           `json:"to_chain"`
	FromChain string           `json:"from_chain"`
	State     string           `json:"state"`
	SendTx    StatusTrackingTx `json:"send_tx"`
	ReceiveTx StatusTrackingTx `json:"receive_tx"`
}

// SkipGoError represents an error response from SkipGo.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

// Error implements the error interface for SkipGoError.
func (e Error) Error() string {
	return fmt.Sprintf("SkipGo Error: Code %d, Message: %s, Details: %+v", e.Code, e.Message, e.Details)
}
