package connection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/margined-protocol/locust-core/pkg/messages/authz"
	"github.com/margined-protocol/locust-core/pkg/types"
	"go.uber.org/zap"

	cometbft "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
)

// ChainMessage wraps a message with its source chain information for proper routing
type ChainMessage struct {
	ChainID     string    // Chain ID where this message should be executed
	Messages    []sdk.Msg // Messages to be executed on the source chain
	IsFeeClient bool      // Whether to use the fee client
	WrapAuthz   bool      // Whether to wrap the messages in an Authz MsgExec
}

// NewAuthzChainMsg creates a new ChainMessage with WrapAuthz set to true
func NewAuthzChainMsg(chainID string, msgs []sdk.Msg) ChainMessage {
	return ChainMessage{
		ChainID:   chainID,
		Messages:  msgs,
		WrapAuthz: true,
	}
}

// NewChainMsg creates a new ChainMessage with WrapAuthz set to false
func NewChainMsg(chainID string, msgs []sdk.Msg) ChainMessage {
	return ChainMessage{
		ChainID:   chainID,
		Messages:  msgs,
		WrapAuthz: false,
	}
}

// MessageSender defines an interface for sending messages using the cosmos client
type MessageSender interface {
	SendAuthzMessages(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) error
	SendAuthzMessagesWithResponse(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) (*cosmosclient.Response, error)
	SendMessages(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) error
	SendMessagesWithResponse(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) (*cosmosclient.Response, error)
}

// DefaultMessageSender is a default implementation of the MessageSender interface
type DefaultMessageSender struct{}

func BroadcastWithRetry(ctx context.Context, l *zap.Logger, cosmosClient *cosmosclient.Client, account cosmosaccount.Account, cfg *types.Config, msgs ...sdk.Msg) error {
	for attempt := 0; attempt < cfg.TxRetryCount; attempt++ {
		l.Debug("Attempting to broadcast transaction",
			zap.Int("attempt", attempt+1),
			zap.Int("max_attempts", cfg.TxRetryCount),
		)

		txResp, err := cosmosClient.BroadcastTx(ctx, account, msgs...)
		if err != nil {
			l.Error("Error broadcasting transaction",
				zap.Int("attempt", attempt+1),
				zap.String("reason", err.Error()),
			)
			continue
		}

		if txResp.Code == 0 {
			l.Info("Transaction successful",
				zap.String("transaction hash", txResp.TxHash),
			)
			return nil
		}

		if txResp.Code == sdkerrors.ErrWrongSequence.ABCICode() {
			l.Debug("Retrying immediately due to sequence number mismatch")
			continue
		}

		l.Error("Transaction failed with code",
			zap.Int("code", int(txResp.Code)),
			zap.String("log", txResp.RawLog),
		)

		if attempt < cfg.TxRetryCount-1 {
			l.Debug("Retrying after delay",
				zap.Duration("retry_delay", cfg.TxRetryDelay),
			)
			time.Sleep(cfg.TxRetryDelay)
		}
	}

	return fmt.Errorf("failed to send transaction after %d attempts", cfg.TxRetryCount)
}

func BroadcastWithRetryAndResponse(ctx context.Context, l *zap.Logger, cosmosClient *cosmosclient.Client, account cosmosaccount.Account, cfg *types.Config, msgs ...sdk.Msg) (*cosmosclient.Response, error) {
	for attempt := 0; attempt < cfg.TxRetryCount; attempt++ {
		l.Debug("Attempting to broadcast transaction",
			zap.Int("attempt", attempt+1),
			zap.Int("max_attempts", cfg.TxRetryCount),
		)

		address, _ := account.Record.GetAddress()
		l.Debug("Broadcasting transaction from address", zap.String("address", address.String()))

		txResp, err := cosmosClient.BroadcastTx(ctx, account, msgs...)
		if err != nil {
			l.Error("Error broadcasting transaction",
				zap.Int("attempt", attempt+1),
				zap.String("reason", err.Error()),
			)
			continue
		}

		if txResp.Code == 0 {
			l.Info("Transaction successful",
				zap.String("transaction hash", txResp.TxHash),
			)
			return &txResp, nil
		}

		if txResp.Code == sdkerrors.ErrWrongSequence.ABCICode() {
			l.Debug("Retrying immediately due to sequence number mismatch")
			continue
		}

		l.Error("Transaction failed with code",
			zap.Int("code", int(txResp.Code)),
			zap.String("log", txResp.RawLog),
		)

		if attempt < cfg.TxRetryCount-1 {
			l.Debug("Retrying after delay",
				zap.Duration("retry_delay", cfg.TxRetryDelay),
			)
			time.Sleep(cfg.TxRetryDelay)
		}
	}

	return nil, fmt.Errorf("failed to send transaction after %d attempts", cfg.TxRetryCount)
}

// BroadcastShortTermOrder broadcasts a transaction and checks for immediate errors,
// specifically designed for dYdX short-term orders which don't get included in blocks
func BroadcastShortTermOrder(ctx context.Context, l *zap.Logger, cosmosClient *cosmosclient.Client, account cosmosaccount.Account, msgs ...sdk.Msg) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	l.Debug("Broadcasting short-term order")

	// Create a channel to signal completion
	done := make(chan error, 1)

	// Broadcast in a goroutine
	go func() {
		txResp, err := cosmosClient.BroadcastTx(ctx, account, msgs...)
		if err != nil {
			done <- fmt.Errorf("failed to broadcast short-term order: %w", err)
			return
		}
		if txResp.Code != 0 {
			done <- fmt.Errorf("order rejected with code %d: %s", txResp.Code, txResp.RawLog)
			return
		}
		done <- nil
	}()

	// Wait for either completion or timeout
	select {
	case err := <-done:
		if err != nil {
			l.Error("Short-term order failed", zap.Error(err))
			return err
		}
		l.Info("Short-term order completed before timeout (unexpected)")
		return nil
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			l.Info("Short-term order broadcast timed out as expected")
			return nil
		}
		l.Error("Context cancelled", zap.Error(ctx.Err()))
		return ctx.Err()
	}
}

// InitRPCClient initialises a RPC client.
func InitRPCClient(logger *zap.Logger, serverAddress, websocketPath string) (*rpchttp.HTTP, cometbft.CometRPC, error) {
	logger.Debug("Initializing RPC client", zap.String("serverAddress", serverAddress), zap.String("websocketPath", websocketPath))
	client, err := rpchttp.New(serverAddress, websocketPath)
	if err != nil {
		logger.Fatal("Error subscribing to websocket client", zap.Error(err))
	}

	logger.Debug("Starting Websocket Client")
	err = client.Start()
	if err != nil {
		logger.Fatal("Error starting websocket client", zap.Error(err))
	}

	return client, client, nil
}

// InitCosmosClient initializes a Cosmos client with retry logic
func InitCosmosClient(ctx context.Context, l *zap.Logger, chain *types.Chain, key *types.SigningKey) (*cosmosclient.Client, error) {
	opts := []cosmosclient.Option{
		cosmosclient.WithNodeAddress(chain.RPCServerAddress),
		cosmosclient.WithAddressPrefix(chain.Prefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringBackend(key.Backend)),
		cosmosclient.WithKeyringDir(key.RootDir),
		cosmosclient.WithKeyringServiceName(key.AppName),
	}

	// If gas, gas adjustment, and gas prices are not set, return an error
	if chain.GasAdjustment == nil && chain.Gas == nil && chain.GasPrices == nil {
		return nil, errors.New("invalid configuration: either fees must be set, or gas, gas adjustment, and gas prices must all be set")
	}

	opts = append(opts,
		cosmosclient.WithGas(*chain.Gas),
		cosmosclient.WithGasAdjustment(*chain.GasAdjustment),
		cosmosclient.WithGasPrices(*chain.GasPrices),
	)

	// Initialise a cosmosclient
	client, err := cosmosclient.New(ctx, opts...)
	if err != nil {
		l.Fatal("Error initialising cosmosclient",
			zap.Error(err),
		)
		return nil, err
	}

	return &client, nil
}

// InitFeeClient initialises a cosmosclient for executing transactions.
func InitFeeClient(ctx context.Context, l *zap.Logger, chain *types.Chain, key *types.SigningKey) (*cosmosclient.Client, error) {
	opts := []cosmosclient.Option{
		cosmosclient.WithNodeAddress(chain.RPCServerAddress),
		cosmosclient.WithAddressPrefix(chain.Prefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringBackend(key.Backend)),
		cosmosclient.WithKeyringDir(key.RootDir),
		cosmosclient.WithKeyringServiceName(key.AppName),
	}

	if chain.Fees == nil {
		return nil, errors.New("invalid configuration: fees must be set")
	}

	opts = append(opts, cosmosclient.WithFees(*chain.Fees))

	// Initialise a cosmosclient
	client, err := cosmosclient.New(ctx, opts...)
	if err != nil {
		l.Fatal("Error initialising cosmosclient",
			zap.Error(err),
		)
		return nil, err
	}

	return &client, nil
}

// InitCosmosQueryClient initialises a cosmosclient for querying which is lightweight and can be disposed of after use.
func InitCosmosQueryClient(ctx context.Context, l *zap.Logger, serverAddress, addressPrefix string) (*cosmosclient.Client, error) {
	opts := []cosmosclient.Option{
		cosmosclient.WithNodeAddress(serverAddress),
		cosmosclient.WithAddressPrefix(addressPrefix),
	}

	// Initialise a cosmosclient
	client, err := cosmosclient.New(ctx, opts...)
	if err != nil {
		l.Fatal("Error initialising cosmosclient",
			zap.Error(err),
		)
		return nil, err
	}

	return &client, nil
}

func GetBlockHeight(ctx context.Context, l *zap.Logger, rpcServerAddress, prefix string) (*int64, error) {
	destinationQueryClient, err := InitCosmosQueryClient(ctx, l, rpcServerAddress, prefix)
	if err != nil {
		l.Fatal("Error creating destination query client", zap.Error(err))
		return nil, err
	}

	status, err := destinationQueryClient.Status(ctx)
	if err != nil {
		l.Fatal("Error getting destination status", zap.Error(err))
		return nil, err
	}

	return &status.SyncInfo.LatestBlockHeight, nil
}

// SendAuthzMessages sends the given messages using the cosmos client and grantee information.
// It returns an error if any step fails.
func (*DefaultMessageSender) SendAuthzMessages(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) error {
	// Get the grantee account and address
	account, granteeAddress, err := GetSignerAccountAndAddress(c, cfg.SignerAccount, cfg.Chain.Prefix)
	if err != nil {
		return fmt.Errorf("error getting grantee account: %w", err)
	}

	l.Info("Grantee address", zap.String("granteeAddress", granteeAddress))

	// Convert the grantee address to AccAddress
	granteeAccAddress, err := sdk.AccAddressFromBech32(granteeAddress)
	if err != nil {
		return fmt.Errorf("failed to generate AccAddress from grantee address: %w", err)
	}

	l.Info("Grantee acc address", zap.String("granteeAccAddress", granteeAccAddress.String()))

	// Create the Authz MsgExec message
	msgExec := authz.CreateAuthzMsg(granteeAccAddress, msgs)

	// Sign and broadcast the message with retries
	err = BroadcastWithRetry(ctx, l, c, *account, cfg, msgExec)
	if err != nil {
		return err
	}

	return nil
}

// SendAuthzMessagesWithResponse sends the given messages using the cosmos client and grantee information.
// It returns an error if any step fails.
func (*DefaultMessageSender) SendAuthzMessagesWithResponse(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) (*cosmosclient.Response, error) {
	// Get the grantee account and address
	account, granteeAddress, err := GetSignerAccountAndAddress(c, cfg.SignerAccount, cfg.Chain.Prefix)
	if err != nil {
		return nil, fmt.Errorf("error getting grantee account: %w", err)
	}

	l.Info("Grantee address", zap.String("granteeAddress", granteeAddress))

	// Convert the grantee address to AccAddress
	granteeAccAddress, err := sdk.AccAddressFromBech32(granteeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AccAddress from grantee address: %w", err)
	}

	l.Info("Grantee acc address", zap.String("granteeAccAddress", granteeAccAddress.String()))

	// Create the Authz MsgExec message
	msgExec := authz.CreateAuthzMsg(granteeAccAddress, msgs)

	// Sign and broadcast the message with retries
	res, err := BroadcastWithRetryAndResponse(ctx, l, c, *account, cfg, msgExec)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SendMessages sends the given messages using the cosmos client and grantee information.
// It returns an error if any step fails.
func (*DefaultMessageSender) SendMessages(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) error {
	// Get the grantee account and address
	account, granteeAddress, err := GetSignerAccountAndAddress(c, cfg.SignerAccount, cfg.Chain.Prefix)
	if err != nil {
		return fmt.Errorf("error getting grantee account: %w", err)
	}

	l.Info("Grantee address", zap.String("granteeAddress", granteeAddress))

	// Convert the grantee address to AccAddress
	granteeAccAddress, err := sdk.AccAddressFromBech32(granteeAddress)
	if err != nil {
		return fmt.Errorf("failed to generate AccAddress from grantee address: %w", err)
	}

	l.Info("Grantee acc address", zap.String("granteeAccAddress", granteeAccAddress.String()))

	// Sign and broadcast the message with retries
	err = BroadcastWithRetry(ctx, l, c, *account, cfg, msgs...)
	if err != nil {
		return err
	}

	return nil
}

// SendMessages sends the given messages using the cosmos client and grantee information.
// It returns an error if any step fails.
func (*DefaultMessageSender) SendMessagesWithResponse(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) (*cosmosclient.Response, error) {
	// Get the grantee account and address
	account, granteeAddress, err := GetSignerAccountAndAddress(c, cfg.SignerAccount, cfg.Chain.Prefix)
	if err != nil {
		return nil, fmt.Errorf("error getting grantee account: %w", err)
	}

	l.Info("Grantee address", zap.String("granteeAddress", granteeAddress))

	// Convert the grantee address to AccAddress
	granteeAccAddress, err := sdk.AccAddressFromBech32(granteeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AccAddress from grantee address: %w", err)
	}

	l.Info("Grantee acc address", zap.String("granteeAccAddress", granteeAccAddress.String()))

	// Sign and broadcast the message with retries
	res, err := BroadcastWithRetryAndResponse(ctx, l, c, *account, cfg, msgs...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SendMessages sends the given messages using the cosmos client and grantee information.
// It returns an error if any step fails.
func (*DefaultMessageSender) SendShortTermMessage(ctx context.Context, l *zap.Logger, c *cosmosclient.Client, cfg *types.Config, msgs ...sdk.Msg) error {
	// Get the grantee account and address

	account, _, err := GetSignerAccountAndAddress(c, cfg.SignerAccount, cfg.Chain.Prefix)
	if err != nil {
		return fmt.Errorf("error getting grantee account: %w", err)
	}

	// Sign and broadcast the message with retries
	err = BroadcastWithRetry(ctx, l, c, *account, cfg, msgs...)
	if err != nil {
		return err
	}

	return nil
}
