package connection

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	proto "github.com/cosmos/gogoproto/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
)

const (
	// Default timeout for gRPC health checks
	DefaultGRPCHealthCheckInterval = 30 * time.Second
	// Default timeout for gRPC connection attempts
	DefaultGRPCConnectionTimeout = 10 * time.Second
)

// GRPCEndpointConfig represents a gRPC endpoint with its TLS configuration
type GRPCEndpointConfig struct {
	Address string
	UseTLS  bool
}

// customCodec implements a custom codec for gogoproto that handles the serialization differences
// between the standard protobuf and gogoproto implementations
type customCodec struct {
	parentCodec encoding.Codec
}

func (c customCodec) Marshal(v interface{}) ([]byte, error) {
	protoMsg, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to assert proto.Message")
	}
	return proto.Marshal(protoMsg)
}

func (c customCodec) Unmarshal(data []byte, v interface{}) error {
	protoMsg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to assert proto.Message")
	}
	return proto.Unmarshal(data, protoMsg)
}

func (c customCodec) Name() string {
	return "gogoproto"
}

// SetupGRPCConnection establishes a GRPC connection, optionally using system's TLS certificates
func SetupGRPCConnection(address string, useTLS bool) (*grpc.ClientConn, error) {
	// Create custom codec for gogoproto compatibility
	customCodec := &customCodec{parentCodec: encoding.GetCodec("proto")}

	var opts []grpc.DialOption

	// Always use the custom codec regardless of TLS setting
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.ForceCodec(customCodec)))

	if useTLS {
		systemCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system certificates: %w", err)
		}

		creds := credentials.NewTLS(&tls.Config{
			RootCAs:    systemCertPool,
			MinVersion: tls.VersionTLS12,
		})

		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", address, err)
	}

	return conn, nil
}

// MultiEndpointGRPCClient manages multiple gRPC endpoints with automatic failover
type MultiEndpointGRPCClient struct {
	ctx               context.Context
	cancel            context.CancelFunc
	logger            *zap.Logger
	mu                sync.RWMutex
	healthCheckTicker *time.Ticker

	// Connection settings
	endpoints           []GRPCEndpointConfig
	currentClient       *grpc.ClientConn
	currentIndex        int
	healthCheckInterval time.Duration
	connectionTimeout   time.Duration
}

// NewMultiEndpointGRPCClient creates a new client that manages multiple gRPC endpoints
func NewMultiEndpointGRPCClient(
	ctx context.Context,
	logger *zap.Logger,
	endpoints []GRPCEndpointConfig,
) (*MultiEndpointGRPCClient, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("at least one endpoint must be provided")
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)

	client := &MultiEndpointGRPCClient{
		ctx:                 ctxWithCancel,
		cancel:              cancel,
		logger:              logger,
		endpoints:           endpoints,
		currentIndex:        0,
		healthCheckInterval: DefaultGRPCHealthCheckInterval,
		connectionTimeout:   DefaultGRPCConnectionTimeout,
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

// Helper function to create a client with addresses and uniform TLS setting
func NewMultiEndpointGRPCClientWithAddresses(
	ctx context.Context,
	logger *zap.Logger,
	addresses []string,
	useTLS bool,
) (*MultiEndpointGRPCClient, error) {
	endpoints := make([]GRPCEndpointConfig, len(addresses))
	for i, addr := range addresses {
		endpoints[i] = GRPCEndpointConfig{
			Address: addr,
			UseTLS:  useTLS,
		}
	}
	return NewMultiEndpointGRPCClient(ctx, logger, endpoints)
}

// connectToEndpoint tries to connect to the current endpoint
func (c *MultiEndpointGRPCClient) connectToEndpoint() error {
	endpoint := c.endpoints[c.currentIndex]
	c.logger.Debug("Connecting to gRPC endpoint",
		zap.String("endpoint", endpoint.Address),
		zap.Bool("useTLS", endpoint.UseTLS))

	// Use the existing SetupGRPCConnection function for consistency
	conn, err := SetupGRPCConnection(endpoint.Address, endpoint.UseTLS)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC endpoint %s: %w", endpoint.Address, err)
	}

	// Update the client reference
	c.mu.Lock()
	if c.currentClient != nil {
		_ = c.currentClient.Close() // Close the old connection if it exists
	}
	c.currentClient = conn
	c.mu.Unlock()

	c.logger.Info("Successfully connected to gRPC endpoint",
		zap.String("endpoint", endpoint.Address),
		zap.Bool("useTLS", endpoint.UseTLS))
	return nil
}

// startHealthCheck begins periodic health checks of the connection
func (c *MultiEndpointGRPCClient) startHealthCheck() {
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
func (c *MultiEndpointGRPCClient) checkConnection() {
	c.mu.RLock()
	conn := c.currentClient
	c.mu.RUnlock()

	if conn == nil {
		// Try to establish initial connection
		c.rotateEndpoint()
		return
	}

	// Check if the connection is in a good state
	state := conn.GetState()
	if state != connectivity.Ready && state != connectivity.Idle {
		c.logger.Warn("gRPC connection is not ready, rotating endpoint",
			zap.String("state", state.String()),
			zap.String("current_endpoint", c.endpoints[c.currentIndex].Address))
		c.rotateEndpoint()
	}
}

// rotateEndpoint switches to the next endpoint in the list and tries to connect
func (c *MultiEndpointGRPCClient) rotateEndpoint() {
	c.mu.Lock()
	// Move to the next endpoint
	c.currentIndex = (c.currentIndex + 1) % len(c.endpoints)
	c.mu.Unlock()

	// Try to connect with the new endpoint
	err := c.connectToEndpoint()
	if err != nil {
		c.logger.Error("Failed to connect to rotated gRPC endpoint",
			zap.Error(err),
			zap.String("endpoint", c.endpoints[c.currentIndex].Address))
	}
}

// GetClient returns the current gRPC client connection
func (c *MultiEndpointGRPCClient) GetClient() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentClient
}

// ForceRotate immediately rotates to the next endpoint
func (c *MultiEndpointGRPCClient) ForceRotate() error {
	c.rotateEndpoint()
	return nil
}

// SetHealthCheckInterval allows changing the interval between health checks
func (c *MultiEndpointGRPCClient) SetHealthCheckInterval(interval time.Duration) {
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
func (c *MultiEndpointGRPCClient) SetConnectionTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connectionTimeout = timeout
}

// GetCurrentEndpoint returns information about the currently connected endpoint
func (c *MultiEndpointGRPCClient) GetCurrentEndpoint() (endpoint GRPCEndpointConfig, index int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.endpoints[c.currentIndex], c.currentIndex
}

// Close shuts down the client and all connections
func (c *MultiEndpointGRPCClient) Close() {
	// Cancel the background goroutine
	c.cancel()

	// Stop the ticker
	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
	}

	// Close connection
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentClient != nil {
		_ = c.currentClient.Close()
		c.currentClient = nil
	}
}
