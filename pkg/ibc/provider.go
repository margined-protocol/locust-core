package ibc

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/margined-protocol/locust-core/pkg/connection"
	"github.com/margined-protocol/locust-core/pkg/utils"
	"go.uber.org/zap"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultTransferProvider implements the TransferProvider interface
type DefaultTransferProvider struct {
	logger *zap.Logger

	// Some housekeeping fields
	baseChainID   string
	signerAccount string

	// Dependencies
	clientRegistry *connection.ClientRegistry
	ibcRegistry    *ConnectionRegistry
	msgHandler     MessageHandler
	listeners      map[string]context.CancelFunc // Websocket listeners
}

// NewTransferProvider creates a new IBC transfer provider
func NewTransferProvider(
	logger *zap.Logger,
	clientRegistry *connection.ClientRegistry,
	baseChainID string,

	signerAccount string,
	msgHandler MessageHandler,
) TransferProvider {
	defaultProvider := &DefaultTransferProvider{
		logger:         logger,
		clientRegistry: clientRegistry,
		baseChainID:    baseChainID,
		signerAccount:  signerAccount,
		ibcRegistry:    DefaultConnectionRegistry(),
		listeners:      make(map[string]context.CancelFunc),
		msgHandler:     msgHandler,
	}

	return defaultProvider
}

// waitForReceivePacket waits for the IBC receive_packet event
func (p *DefaultTransferProvider) waitForReceivePacket(
	ctx context.Context,
	request *TransferRequest,
) error {
	// Create a channel for the balance check result
	balanceCh := make(chan error, 1)

	// Start balance polling in a separate goroutine
	go func() {
		// Get initial balance
		initialBalance, err := p.queryBalance(ctx, request.DestinationChain, request.Receiver, request.RecvDenom)
		if err != nil {
			p.logger.Warn("Failed to query initial balance",
				zap.Error(err),
				zap.String("chain", request.DestinationChain),
				zap.String("receiver", request.Receiver))
			balanceCh <- err
			return
		}

		// Poll balance every 10 seconds
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		expectedAmount := request.Amount.Amount
		for {
			select {
			case <-ctx.Done():
				balanceCh <- ctx.Err()
				return
			case <-ticker.C:
				currentBalance, err := p.queryBalance(ctx, request.DestinationChain, request.Receiver, request.RecvDenom)
				if err != nil {
					continue // Skip this iteration if query fails
				}

				// Calculate balance difference
				diff := currentBalance.Amount.Sub(initialBalance.Amount)

				// Check if we received at least 95% of expected amount (accounting for fees)
				minExpectedAmount := expectedAmount.Mul(sdkmath.NewInt(95)).Quo(sdkmath.NewInt(100))
				if diff.GTE(minExpectedAmount) {
					p.logger.Info("Detected balance increase",
						zap.String("initial", initialBalance.String()),
						zap.String("current", currentBalance.String()),
						zap.String("difference", diff.String()),
						zap.String("expected", expectedAmount.String()))
					balanceCh <- nil
					return
				}
			}
		}
	}()

	// Get the destination chain's RPC endpoint
	destClientInstance, err := p.clientRegistry.GetClient(request.DestinationChain, false)
	if err != nil {
		return fmt.Errorf("failed to get destination chain client: %w", err)
	}

	// Create websocket client
	wsClient, _, err := connection.InitRPCClient(p.logger, destClientInstance.Chain.RPCServerAddress, "/websocket")
	if err != nil {
		return fmt.Errorf("failed to create websocket client: %w", err)
	}
	defer func() {
		if err := wsClient.Stop(); err != nil {
			p.logger.Error("failed to stop websocket client", zap.Error(err))
		}
	}()

	// Subscribe to receive_packet events
	query := fmt.Sprintf("transfer.recipient = '%s'", request.Receiver)
	eventCh := utils.CreateEventChannel(ctx, p.logger, wsClient, query)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case event := <-eventCh:
		p.logger.Info("Received IBC packet",
			zap.Any("event", event.Query)) // Note: don't print the whole event, it's too big
		return nil
	case err := <-balanceCh:
		if err != nil {
			return fmt.Errorf("balance polling failed: %w", err)
		}
		p.logger.Info("Transfer confirmed via balance change")
		return nil
	}
}

// Transfer initiates an IBC transfer between chains
func (p *DefaultTransferProvider) Transfer(ctx context.Context, request *TransferRequest) (*TransferResult, error) {
	// Clear any previous instances
	runtime.GC()

	p.logger.Info("Transferring", zap.Any("request", request))
	result := TransferResult{}

	// Get the clients for source and destination chains
	sourceClientInstance, err := p.clientRegistry.GetClient(request.SourceChain, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get source chain client: %w", err)
	}

	// Assign receiver address for non-base chains
	if request.DestinationChain != p.baseChainID {
		_, receiver, err := p.clientRegistry.GetSignerAccountAndAddress(p.signerAccount, request.DestinationChain)
		if err != nil {
			p.logger.Error("Failed to get address for chain", zap.String("chain", request.DestinationChain), zap.Error(err))
			return nil, err
		}

		request.Receiver = receiver
	}

	// Get source balance before transfer
	sourceBalance, err := p.queryBalance(ctx, request.SourceChain, request.Sender, request.Amount.Denom)
	if err != nil {
		p.logger.Warn("Failed to query source balance before transfer",
			zap.Error(err),
			zap.String("chain", request.SourceChain),
			zap.String("address", request.Sender),
			zap.String("denom", request.Amount.Denom),
		)
	}
	p.logger.Info("Source balance", zap.Any("balance", sourceBalance))

	// Get destination balance before transfer
	destDenom := request.RecvDenom
	if destDenom == "" {
		// If no receive denom specified, use same denom (though it may be prefixed with ibc/ on destination)
		destDenom = request.Amount.Denom
	}

	destBalance, err := p.queryBalance(ctx, request.DestinationChain, request.Receiver, destDenom)
	if err != nil {
		p.logger.Warn("Failed to query destination balance before transfer",
			zap.Error(err),
			zap.String("chain", request.DestinationChain),
			zap.String("address", request.Receiver),
			zap.String("denom", destDenom),
		)
	}
	p.logger.Info("Destination balance", zap.Any("balance", destBalance))

	// Get the IBC connection for this transfer
	conn, err := p.ibcRegistry.GetConnection(request.SourceChain, request.DestinationChain)
	if err != nil {
		return nil, fmt.Errorf("failed to get IBC connection: %w", err)
	}

	chainID := request.SourceChain
	if conn.Transfer.Forward != nil {
		// Get account address for this chain
		_, forwardReceiver, err := p.clientRegistry.GetSignerAccountAndAddress(p.signerAccount, conn.Transfer.Forward.ChainID)
		if err != nil {
			p.logger.Error("Failed to get address for chain", zap.String("chain", conn.Transfer.Forward.ChainID), zap.Error(err))
			return nil, err
		}

		chainID = conn.Transfer.Forward.ChainID
		conn.Transfer.Forward.Receiver = forwardReceiver
	}

	// Get source chain block height
	blockHeight, err := p.clientRegistry.GetHeight(ctx, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source chain height: %w", err)
	}

	// Prepare timeout height
	// nolint
	timeout := uint64(*blockHeight) + request.Timeout

	// Get the account and sender address with correct prefix
	account, sender, err := connection.GetSignerAccountAndAddress(
		sourceClientInstance.Client,
		p.signerAccount,
		sourceClientInstance.Chain.Prefix,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get signer account and address: %w", err)
	}

	sdkaddr, _ := account.Record.GetAddress()
	p.logger.Info("Sender", zap.String("sender", sender))
	p.logger.Info("Account", zap.String("account", sdkaddr.String()))

	// Create transfer message
	transferMsg, err := CreateTransferWithMemo(
		conn.Transfer,
		request.SourceChain,
		request.DestinationChain,
		request.Amount,
		timeout,
		request.Sender,
		request.Receiver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer message: %w", err)
	}

	p.logger.Info("Transfer message", zap.Any("message", transferMsg))

	// Send the message
	p.logger.Info("Sending IBC transfer",
		zap.String("source", request.SourceChain),
		zap.String("destination", request.DestinationChain),
		zap.String("amount", request.Amount.String()),
	)

	response, err := p.msgHandler(request.SourceChain, []sdk.Msg{transferMsg}, false, true)
	if err != nil {
		return &TransferResult{Error: err}, err
	}

	// Update transfer data with source transaction result
	result.SourceTxHash = response.TxHash
	result.SourceResponse = response

	// If requested, wait for the receive packet event
	// Create a context with timeout if specified
	waitCtx, cancel := context.WithTimeout(ctx, 6*time.Minute)
	defer cancel()

	err = p.waitForReceivePacket(waitCtx, request)
	if err != nil {
		p.logger.Warn("Failed to wait for receive packet", zap.Error(err))
	}

	return &result, nil
}

// queryBalance queries the balance of an address on a specific chain
func (p *DefaultTransferProvider) queryBalance(ctx context.Context, chainID, address, denom string) (sdk.Coin, error) {
	client, err := p.clientRegistry.GetClient(chainID, false)
	if err != nil {
		return sdk.Coin{}, err
	}

	coins, err := client.Client.BankBalances(ctx, address, nil)
	if err != nil {
		return sdk.Coin{}, err
	}

	for _, coin := range coins {
		if coin.Denom == denom {
			return coin, nil
		}
	}

	// Return zero amount if denom not found
	return sdk.Coin{}, nil
}
