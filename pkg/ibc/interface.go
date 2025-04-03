package ibc

import (
	"context"
	"time"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TransferProvider defines the interface for IBC transfers between chains
type TransferProvider interface {
	// Transfer initiates an IBC transfer between chains.
	// If the WaitForCompletion flag is set to true in the request, this method will block
	// until the transfer is completed or the timeout is reached.
	Transfer(ctx context.Context, request *TransferRequest) (*TransferResult, error)
}

// MessageHandler is a callback function to execute messages
type MessageHandler func(chainID string, msgs []sdk.Msg, isFeeClient bool, wrapAuthz bool) (*cosmosclient.Response, error)

// TransferRequest contains all parameters needed for an IBC transfer
type TransferRequest struct {
	SourceChain       string        // Source chain ID
	DestinationChain  string        // Destination chain ID
	Sender            string        // Sender address on source chain
	Receiver          string        // Receiver address on destination chain
	Amount            sdk.Coin      // Amount to transfer
	RecvDenom         string        // Denom of the token to receive
	Timeout           uint64        // Timeout in blocks
	Fee               sdk.Coins     // Optional fee for the transfer
	CompletionTimeout time.Duration // Maximum time to wait for transfer completion
}

// TransferResult contains the result of a transfer operation
type TransferResult struct {
	SourceTxHash   string                 // Hash of the source chain transaction
	DestTxHash     string                 // Hash of the destination chain transaction if available
	Error          error                  // Error if transfer failed
	SourceResponse *cosmosclient.Response // Source chain response
	DestResponse   *cosmosclient.Response // Destination chain response (if available)
}
