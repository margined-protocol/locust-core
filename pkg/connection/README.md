# Connection Package

This package provides connection management for different network protocols (RPC, gRPC) used in the Locust project.

## Recent Changes

The gRPC functionality previously in the separate `grpc` package has been consolidated into this package to:

1. Reduce code duplication
2. Provide a unified interface for all connection types
3. Simplify maintenance of connection management logic
4. Standardize how we handle connection failover and health checking

## Components

### RPC Clients

- `MultiEndpointRPCClient`: Manages multiple RPC endpoints with automatic failover
- `InitRPCClient`: Creates a single RPC client to a specific endpoint

### gRPC Clients

- `MultiEndpointGRPCClient`: Manages multiple gRPC endpoints with automatic failover
- `SetupGRPCConnection`: Creates a single gRPC connection to a specific endpoint

### Message Management

- `MessageSender`: Interface for sending messages to the blockchain
- `DefaultMessageSender`: Default implementation of the MessageSender interface
- Various helper functions for broadcasting transactions with retry logic

### Client Registry

- `ClientRegistry`: Centralized registry of Cosmos clients for different chains
- Methods for initializing, retrieving, and managing client instances

## Migration from grpc package

If you were previously using the `grpc` package directly, you should update your imports to use the `connection` package instead:

```go
// Before
import "github.com/margined-protocol/locust-core/pkg/grpc"

// After
import "github.com/margined-protocol/locust-core/pkg/connection"
```

Function and type names remain the same, so code should work with minimal changes beyond the import path update.
