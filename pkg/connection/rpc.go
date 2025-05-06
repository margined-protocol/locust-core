package connection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	cometbft "github.com/cosmos/cosmos-sdk/client"
)

const (
	// Default timeout for RPC health checks
	DefaultRPCHealthCheckInterval = 30 * time.Second
	// Default timeout for RPC connection attempts
	DefaultRPCConnectionTimeout = 10 * time.Second
)

// RPCEndpointConfig represents an RPC endpoint configuration
type RPCEndpointConfig struct {
	Address       string
	WebsocketPath string
}

// MultiEndpointRPCClient manages multiple RPC endpoints with automatic failover
type MultiEndpointRPCClient struct {
	ctx               context.Context
	cancel            context.CancelFunc
	logger            *zap.Logger
	mu                sync.RWMutex
	healthCheckTicker *time.Ticker

	// Connection settings
	endpoints           []RPCEndpointConfig
	currentClient       *rpchttp.HTTP
	cometClient         cometbft.CometRPC // Interface used by many cosmos functions
	currentIndex        int
	healthCheckInterval time.Duration
	connectionTimeout   time.Duration
}

// NewMultiEndpointRPCClient creates a new client that manages multiple RPC endpoints
func NewMultiEndpointRPCClient(
	ctx context.Context,
	logger *zap.Logger,
	endpoints []RPCEndpointConfig,
) (*MultiEndpointRPCClient, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("at least one endpoint must be provided")
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)

	client := &MultiEndpointRPCClient{
		ctx:                 ctxWithCancel,
		cancel:              cancel,
		logger:              logger,
		endpoints:           endpoints,
		currentIndex:        0,
		healthCheckInterval: DefaultRPCHealthCheckInterval,
		connectionTimeout:   DefaultRPCConnectionTimeout,
	}

	// Try to establish initial connection
	err := client.connectToEndpoint()
	if err != nil {
		cancel() // Clean up context if connection fails
		return nil, fmt.Errorf("failed to connect to any endpoint: %w", err)
	}

	// Start health check routine
	client.startHealthCheck()

	return client, nil
}

// Helper function to create a client with addresses and the same websocket path
func NewMultiEndpointRPCClientWithAddresses(
	ctx context.Context,
	logger *zap.Logger,
	addresses []string,
	websocketPath string,
) (*MultiEndpointRPCClient, error) {
	endpoints := make([]RPCEndpointConfig, len(addresses))
	for i, addr := range addresses {
		endpoints[i] = RPCEndpointConfig{
			Address:       addr,
			WebsocketPath: websocketPath,
		}
	}
	return NewMultiEndpointRPCClient(ctx, logger, endpoints)
}

// connectToEndpoint tries to connect to the current endpoint
func (c *MultiEndpointRPCClient) connectToEndpoint() error {
	endpoint := c.endpoints[c.currentIndex]
	c.logger.Debug("Connecting to RPC endpoint",
		zap.String("address", endpoint.Address),
		zap.String("websocket_path", endpoint.WebsocketPath))

	// Use the existing InitRPCClient function for consistency
	client, cometClient, err := InitRPCClient(c.logger, endpoint.Address, endpoint.WebsocketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC endpoint %s: %w", endpoint.Address, err)
	}

	// Update the client reference
	c.mu.Lock()
	c.currentClient = client
	c.cometClient = cometClient
	c.mu.Unlock()

	c.logger.Info("Successfully connected to RPC endpoint",
		zap.String("address", endpoint.Address))
	return nil
}

// startHealthCheck begins periodic health checks of the connection
func (c *MultiEndpointRPCClient) startHealthCheck() {
	c.healthCheckTicker = time.NewTicker(c.healthCheckInterval)

	go func() {
		for {
			select {
			case <-c.healthCheckTicker.C:
				c.checkConnection()
			case <-c.ctx.Done():
				if c.healthCheckTicker != nil {
					c.healthCheckTicker.Stop()
				}
				return
			}
		}
	}()
}

// checkConnection verifies that the connection is healthy and rotates to next endpoint if needed
func (c *MultiEndpointRPCClient) checkConnection() {
	c.mu.RLock()
	client := c.currentClient
	c.mu.RUnlock()

	if client == nil {
		// Try to establish initial connection
		c.rotateEndpoint()
		return
	}

	// Create context with timeout for health check
	ctx, cancel := context.WithTimeout(c.ctx, c.connectionTimeout)
	defer cancel()

	// Check if the connection is still alive by making a Status request
	_, err := client.Status(ctx)
	if err != nil {
		c.logger.Warn("RPC connection health check failed, rotating endpoint",
			zap.Error(err),
			zap.String("current_endpoint", c.endpoints[c.currentIndex].Address))
		c.rotateEndpoint()
	}
}

// rotateEndpoint switches to the next endpoint in the list and tries to connect
func (c *MultiEndpointRPCClient) rotateEndpoint() {
	c.mu.Lock()
	// Move to the next endpoint
	c.currentIndex = (c.currentIndex + 1) % len(c.endpoints)
	c.mu.Unlock()

	// Try to connect with the new endpoint
	err := c.connectToEndpoint()
	if err != nil {
		c.logger.Error("Failed to connect to rotated RPC endpoint",
			zap.Error(err),
			zap.String("endpoint", c.endpoints[c.currentIndex].Address))
	}
}

// GetClient returns the current RPC client
func (c *MultiEndpointRPCClient) GetClient() *rpchttp.HTTP {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentClient
}

// GetCometClient returns the current CometRPC client interface
func (c *MultiEndpointRPCClient) GetCometClient() cometbft.CometRPC {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cometClient
}

// ForceRotate immediately rotates to the next endpoint
func (c *MultiEndpointRPCClient) ForceRotate() error {
	c.rotateEndpoint()
	return nil
}

// SetHealthCheckInterval allows changing the interval between health checks
func (c *MultiEndpointRPCClient) SetHealthCheckInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.healthCheckInterval = interval

	// Restart health check with new interval if running
	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
		c.healthCheckTicker = time.NewTicker(c.healthCheckInterval)
	}
}

// SetConnectionTimeout sets the timeout for connection attempts
func (c *MultiEndpointRPCClient) SetConnectionTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connectionTimeout = timeout
}

// GetCurrentEndpoint returns information about the currently connected endpoint
func (c *MultiEndpointRPCClient) GetCurrentEndpoint() (endpoint RPCEndpointConfig, index int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.endpoints[c.currentIndex], c.currentIndex
}

// Close shuts down the client and all connections
func (c *MultiEndpointRPCClient) Close() {
	// Cancel the background goroutine
	c.cancel()

	// Stop the ticker
	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
	}

	// Close connection (not needed for RPC client as it doesn't expose a Close method)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Though the HTTP client doesn't have a Close method, we should stop the websocket connection
	if c.currentClient != nil {
		_ = c.currentClient.Stop()
		c.currentClient = nil
		c.cometClient = nil
	}
}
