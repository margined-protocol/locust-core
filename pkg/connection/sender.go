package connection

import (
	"context"
	"fmt"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/margined-protocol/locust-core/pkg/types"
	"go.uber.org/zap"
)

func SendMessages(
	ctx context.Context, l *zap.Logger,
	clientRegistry *ClientRegistry,
	chainMsg ChainMessage,
	messageSender MessageSender,
	cfg *types.Config,
	isDryRun, isFeeClient, wrapAuthz bool,
) (*cosmosclient.Response, error) {
	if isDryRun {
		l.Debug("Dry run enabled", zap.Any("Generated messages", chainMsg.Messages))
		return nil, nil
	}

	if len(chainMsg.Messages) == 0 {
		l.Info("No messages to send")
		return nil, nil
	}

	l.Info("Sending messages", zap.Any("messages", chainMsg.Messages))
	l.Info("Chain ID", zap.String("chainID", chainMsg.ChainID))
	chain, err := clientRegistry.GetClient(chainMsg.ChainID, isFeeClient)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	l.Info("Chain", zap.Any("chain", chain.Chain))

	tmpCfg := types.Config{
		Chain:         *chain.Chain,
		SignerAccount: cfg.SignerAccount,
		DryRun:        isDryRun,
		TxRetryCount:  cfg.TxRetryCount,
		TxRetryDelay:  cfg.TxRetryDelay,
	}

	msgs := chainMsg.Messages

	var resp *cosmosclient.Response
	if wrapAuthz {
		resp, err = messageSender.SendAuthzMessagesWithResponse(ctx, l, chain.Client, &tmpCfg, msgs...)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err = messageSender.SendMessagesWithResponse(ctx, l, chain.Client, &tmpCfg, msgs...)
		if err != nil {
			return nil, err
		}
	}

	// Wait for transaction confirmation
	l.Debug("Waiting for transaction confirmation", zap.String("txHash", resp.TxHash))
	txResult, err := chain.Client.WaitForTx(ctx, resp.TxHash)
	if err != nil {
		return nil, fmt.Errorf("error waiting for transaction confirmation: %w", err)
	}

	l.Debug(
		"Transaction confirmed",
		zap.String("TxHash", resp.TxHash),
		zap.String("BlockHash", fmt.Sprintf("%X", txResult.Hash)),
	)

	return resp, nil
}
