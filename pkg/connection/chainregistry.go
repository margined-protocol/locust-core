package connection

import (
	"context"
	"fmt"
	"sync"

	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/margined-protocol/locust-core/pkg/grpc"
	"github.com/margined-protocol/locust-core/pkg/types"
	"github.com/margined-protocol/locust-core/pkg/utils"
	"go.uber.org/zap"

	sdkmath "cosmossdk.io/math"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// ClientEntry represents a registered chain with its client and configuration
type ClientEntry struct {
	Chain *types.Chain
	Key   *types.SigningKey
}

// ClientInstance represents a registered chain with its client and configuration
type ClientInstance struct {
	Client *cosmosclient.Client
	Chain  *types.Chain
	Key    *types.SigningKey
}

// ClientRegistry manages connections to multiple chains
type ClientRegistry struct {
	chains        map[string]*ClientEntry
	logger        *zap.Logger
	mu            sync.RWMutex
	signerAccount string
}

// NewClientRegistry creates a new chain registry
func NewClientRegistry(logger *zap.Logger, signerAccount string) *ClientRegistry {
	return &ClientRegistry{
		chains:        make(map[string]*ClientEntry),
		logger:        logger,
		signerAccount: signerAccount,
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

	r.mu.Lock()
	r.chains[chain.ChainID] = &ClientEntry{
		Chain: chain,
		Key:   chainKey,
	}
	r.mu.Unlock()

	r.logger.Info("Client registered successfully",
		zap.String("chain_id", chain.ChainID),
		zap.String("rpc_address", chain.RPCServerAddress))

	return nil
}

// GetClient retrieves a chain client entry by its ID
func (r *ClientRegistry) initClient(chainID string, isFee bool) (*ClientInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chain, exists := r.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain %s not registered", chainID)
	}

	// Initialize Cosmos client
	var client *cosmosclient.Client
	var err error
	if isFee {
		client, err = InitFeeClient(context.Background(), r.logger, chain.Chain, chain.Key)
	} else {
		client, err = InitCosmosClient(context.Background(), r.logger, chain.Chain, chain.Key)
	}
	if err != nil {
		r.logger.Fatal("Error initializing cosmos client", zap.Error(err))
	}

	return &ClientInstance{
		Client: client,
		Chain:  chain.Chain,
		Key:    chain.Key,
	}, nil
}

// GetClient retrieves a chain client entry by its ID
func (r *ClientRegistry) GetClient(chainID string, isFeeClient bool) (*ClientInstance, error) {
	return r.initClient(chainID, isFeeClient)
}

// Getheight retrieves the blockheigh of a chain by its ID
func (r *ClientRegistry) GetHeight(ctx context.Context, chainID string) (*int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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
	defer r.mu.RUnlock()

	client, err := r.GetClient(chainID, false)
	if err != nil {
		return nil, err
	}

	// Initialize GRPC connections
	conn, err := grpc.SetupGRPCConnection(client.Chain.GRPCServerAddress, client.Chain.GRPCTLS)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GRPC server: %w", err)
	}

	bankClient := banktypes.NewQueryClient(conn)

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

// Close closes all clients
func (r *ClientRegistry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Close any resources that need closing
	r.chains = make(map[string]*ClientEntry)
	r.logger.Info("Client registry closed")
}
