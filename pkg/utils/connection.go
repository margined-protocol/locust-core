package utils

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
)

func CreateEventChannel(ctx context.Context, l *zap.Logger, ws *rpchttp.HTTP, wsQuery string) <-chan coretypes.ResultEvent {
	l.Debug("Websocket Query",
		zap.String("Generated query", wsQuery),
	)

	// Generate a UUID for the websocket client
	id, err := uuid.NewUUID()
	if err != nil {
		l.Fatal("Error creating New UUID", zap.Error(err))
	}

	// Generate the unique string
	subscriber := "locust-" + id.String()
	l.Debug("Websocket subscriber string", zap.String("subscriber", subscriber))

	// Listen on the websocket connection for the query
	eventChSwap, err := ws.Subscribe(ctx, subscriber, wsQuery)
	if err != nil {
		l.Fatal("Error subscribing websocket client", zap.Error(err))
	}

	return eventChSwap
}
