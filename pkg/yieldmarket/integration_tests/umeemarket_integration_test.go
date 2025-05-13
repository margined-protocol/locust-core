// locust-core/pkg/yieldmarket/integration_tests/umeemarket_integration_test.go

package yieldmarket

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/margined-protocol/locust-core/pkg/yieldmarket"
	"github.com/test-go/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var runIntegrationTests = flag.Bool("integration", false, "run integration tests")

func TestRefreshMarketData(t *testing.T) {
	if !*runIntegrationTests {
		t.Skip("skipping integration test")
	}

	// Set up a real gRPC connection to the Umee market
	conn, err := grpc.Dial("umee-grpc.polkachu.com:13690", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to connect to Umee market: %v", err)
	}
	defer conn.Close()

	logger, _ := zap.NewProduction()

	// Call the standalone function
	marketData, err := yieldmarket.RefreshMarketData(
		context.Background(),
		conn,
		time.Now(),
		"ibc/92BC8E5C50E6664B4DA748B62C1FFBE321967E1F8868EE03B005977F9AA7C0B8",
		logger,
	)
	if err != nil {
		t.Fatalf("failed to refresh market data: %v", err)
	}

	// Add assertions to verify the behavior of the function
	if marketData == nil {
		t.Fatalf("expected market data, got nil")
	}
	t.Logf("Market Data: %v", marketData)

	assert.Equal(t, "123", "uumee")
}
