package connection

import (
	"context"
	"fmt"
	"sync"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/margined-protocol/locust-core/pkg/types"
	"github.com/margined-protocol/locust-core/pkg/utils"
	"go.uber.org/zap"

	sdkmath "cosmossdk.io/math"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// ClientEntry represents a registered chain with its client and configuration
type ClientEntry struct {
	Chain      *types.Chain
	Key        *types.SigningKey
	GRPCClient *MultiEndpointGRPCClient
	RPCClient  *MultiEndpointRPCClient
}

// ClientInstance represents a registered chain with its client and configuration
type ClientInstance struct {
	Client     *cosmosclient.Client
	Chain      *types.Chain
	Key        *types.SigningKey
	GRPCClient *MultiEndpointGRPCClient
	RPCClient  *MultiEndpointRPCClient
}

// ClientRegistry manages connections to multiple chains
type ClientRegistry struct {
	chains        map[string]*ClientEntry
	logger        *zap.Logger
	mu            sync.RWMutex
	signerAccount string
	ctx           context.Context
}

// NewClientRegistry creates a new chain registry
func NewClientRegistry(ctx context.Context, logger *zap.Logger, signerAccount string) *ClientRegistry {
	return &ClientRegistry{
		chains:        make(map[string]*ClientEntry),
		logger:        logger,
		signerAccount: signerAccount,
		ctx:           ctx,
	}
}

// RegisterClient adds a new chain client to the registry
func (r *ClientRegistry) RegisterClient(chain *types.Chain, key *types.SigningKey) error {
	if _, exists := r.chains[chain.ChainID]; exists {
		return fmt.Errorf("chain %s already registered", chain.ChainID)
	}

	// Create a new key instance for this specific chain
	chainKey := &types.SigningKey{
		AppName: key.AppName,
		Backend: key.Backend,
		RootDir: key.RootDir,
	}

	// Create multi-endpoint gRPC client
	var grpcClient *MultiEndpointGRPCClient
	if len(chain.GRPCEndpoints) > 0 {
		// Create endpoints configuration
		endpoints := make([]GRPCEndpointConfig, len(chain.GRPCEndpoints))
		for i, endpoint := range chain.GRPCEndpoints {
			endpoints[i] = GRPCEndpointConfig{
				Address: endpoint.Address,
				UseTLS:  endpoint.UseTLS,
			}
		}

		// Initialize the gRPC client
		var err error
		grpcClient, err = NewMultiEndpointGRPCClient(r.ctx, r.logger, endpoints)
		if err != nil {
			return fmt.Errorf("failed to initialize gRPC client: %w", err)
		}
	}

	// Create multi-endpoint RPC client
	var rpcClient *MultiEndpointRPCClient
	if len(chain.RPCEndpoints) > 0 {
		// Create endpoints configuration
		endpoints := make([]RPCEndpointConfig, len(chain.RPCEndpoints))
		for i, endpoint := range chain.RPCEndpoints {
			endpoints[i] = RPCEndpointConfig{
				Address:       endpoint.Address,
				WebsocketPath: "/websocket", // Default value, could be configurable
			}
		}

		// Initialize the RPC client
		var err error
		rpcClient, err = NewMultiEndpointRPCClient(r.ctx, r.logger, endpoints)
		if err != nil {
			// Clean up gRPC client if it was created
			if grpcClient != nil {
				grpcClient.Close()
			}
			return fmt.Errorf("failed to initialize RPC client: %w", err)
		}
	}

	r.mu.Lock()
	r.chains[chain.ChainID] = &ClientEntry{
		Chain:      chain,
		Key:        chainKey,
		GRPCClient: grpcClient,
		RPCClient:  rpcClient,
	}
	r.mu.Unlock()

	r.logger.Info("Client registered successfully",
		zap.String("chain_id", chain.ChainID),
		zap.Int("grpc_endpoints", len(chain.GRPCEndpoints)),
		zap.Int("rpc_endpoints", len(chain.RPCEndpoints)))

	return nil
}

// GetClient retrieves a chain client entry by its ID
func (r *ClientRegistry) initClient(chainID string, isFee bool) (*ClientInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain %s not registered", chainID)
	}

	// Initialize Cosmos client
	var client *cosmosclient.Client
	var err error

	// Override cosmosclient options with the current active RPC endpoint
	rpcAddress := ""
	if entry.RPCClient != nil {
		currentEndpoint, _ := entry.RPCClient.GetCurrentEndpoint()
		rpcAddress = currentEndpoint.Address
	} else if len(entry.Chain.RPCEndpoints) > 0 {
		// Fallback to first endpoint if multi-client not initialized
		rpcAddress = entry.Chain.RPCEndpoints[0].Address
	}

	if isFee {
		// Create a temporary chain with the active RPC endpoint
		tempChain := *entry.Chain // Create a copy
		if rpcAddress != "" {
			// Use the active RPC endpoint
			tempChain.RPCEndpoints[0].Address = rpcAddress
		}
		client, err = InitFeeClient(context.Background(), r.logger, &tempChain, entry.Key)
	} else {
		// Create a temporary chain with the active RPC endpoint
		tempChain := *entry.Chain // Create a copy
		if rpcAddress != "" {
			// Use the active RPC endpoint
			tempChain.RPCEndpoints[0].Address = rpcAddress
		}
		client, err = InitCosmosClient(context.Background(), r.logger, &tempChain, entry.Key)
	}
	if err != nil {
		r.logger.Error("Error initializing cosmos client", zap.Error(err))
		return nil, err
	}

	return &ClientInstance{
		Client:     client,
		Chain:      entry.Chain,
		Key:        entry.Key,
		GRPCClient: entry.GRPCClient,
		RPCClient:  entry.RPCClient,
	}, nil
}

// GetClient retrieves a chain client entry by its ID
func (r *ClientRegistry) GetClient(chainID string, isFeeClient bool) (*ClientInstance, error) {
	return r.initClient(chainID, isFeeClient)
}

// GetHeight retrieves the blockheight of a chain by its ID
func (r *ClientRegistry) GetHeight(ctx context.Context, chainID string) (*int64, error) {
	r.mu.RLock()
	entry, exists := r.chains[chainID]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("chain %s not registered", chainID)
	}

	// Use the multi-endpoint RPC client if available
	if entry.RPCClient != nil {
		rpcClient := entry.RPCClient.GetClient()
		status, err := rpcClient.Status(ctx)
		if err != nil {
			return nil, err
		}
		return &status.SyncInfo.LatestBlockHeight, nil
	}

	// Fallback to the standard client
	client, err := r.GetClient(chainID, false)
	if err != nil {
		return nil, err
	}

	status, err := client.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	return &status.SyncInfo.LatestBlockHeight, nil
}

// GetBalance retrieves the balance of an account on a chain by its ID
func (r *ClientRegistry) GetBalance(ctx context.Context, chainID, signerAccount, denom string) (*sdkmath.Int, error) {
	// This needs to come before the mutex lock else it will block
	_, sender, err := r.GetSignerAccountAndAddress(signerAccount, chainID)
	if err != nil {
		return nil, err
	}

	r.mu.RLock()
	entry, exists := r.chains[chainID]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("chain %s not registered", chainID)
	}

	// Use the multi-endpoint gRPC client if available
	var bankClient banktypes.QueryClient
	if entry.GRPCClient != nil {
		grpcConn := entry.GRPCClient.GetClient()
		bankClient = banktypes.NewQueryClient(grpcConn)
	} else {
		// Fallback to standard client
		client, err := r.GetClient(chainID, false)
		if err != nil {
			return nil, err
		}

		// Get the first gRPC endpoint if available
		if len(client.Chain.GRPCEndpoints) == 0 {
			return nil, fmt.Errorf("no gRPC endpoints configured for chain %s", chainID)
		}

		// Use the first endpoint (fallback)
		conn, err := SetupGRPCConnection(
			client.Chain.GRPCEndpoints[0].Address,
			client.Chain.GRPCEndpoints[0].UseTLS,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to GRPC server: %w", err)
		}

		bankClient = banktypes.NewQueryClient(conn)
	}

	balance, err := utils.GetBalance(ctx, bankClient, sender, denom)
	if err != nil {
		return nil, err
	}

	amount, ok := sdkmath.NewIntFromString(balance.Balance.Amount.String())
	if !ok {
		return nil, fmt.Errorf("failed to parse balance amount: %s", balance.Balance.Amount)
	}

	return &amount, nil
}

// GetSignerAccountAndAddress retrieves the account and address for a specific chain
func (r *ClientRegistry) GetSignerAccountAndAddress(signerAccount, chainID string) (*cosmosaccount.Account, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, err := r.GetClient(chainID, false)
	if err != nil {
		return nil, "", fmt.Errorf("error getting client: %w", err)
	}

	r.logger.Info("Getting signer account and address",
		zap.String("chain_id", chainID),
		zap.String("signer_account", signerAccount),
		zap.String("prefix", client.Chain.Prefix),
	)

	account, sender, err := GetSignerAccountAndAddress(client.Client, signerAccount, client.Chain.Prefix)
	if err != nil {
		return nil, "", fmt.Errorf("error getting signer account and address: %w", err)
	}

	return account, sender, nil
}

// HasClient checks if a client is registered
func (r *ClientRegistry) HasClient(chainID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.chains[chainID]
	return exists
}

// GetGRPCClient returns the multi-endpoint gRPC client for a chain
func (r *ClientRegistry) GetGRPCClient(chainID string) (*MultiEndpointGRPCClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain %s not registered", chainID)
	}

	if entry.GRPCClient == nil {
		return nil, fmt.Errorf("no gRPC client initialized for chain %s", chainID)
	}

	return entry.GRPCClient, nil
}

// GetRPCClient returns the multi-endpoint RPC client for a chain
func (r *ClientRegistry) GetRPCClient(chainID string) (*MultiEndpointRPCClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain %s not registered", chainID)
	}

	if entry.RPCClient == nil {
		return nil, fmt.Errorf("no RPC client initialized for chain %s", chainID)
	}

	return entry.RPCClient, nil
}

// Close closes all clients
func (r *ClientRegistry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Close all multi-endpoint clients
	for _, entry := range r.chains {
		if entry.GRPCClient != nil {
			entry.GRPCClient.Close()
		}
		if entry.RPCClient != nil {
			entry.RPCClient.Close()
		}
	}

	// Reset the registry
	r.chains = make(map[string]*ClientEntry)
	r.logger.Info("Client registry closed")
}
