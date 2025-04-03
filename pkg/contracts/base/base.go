package base

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc"
)

// ContractQueryClient defines the interface for querying contract states
type ContractQueryClient interface {
	QueryRawContractState(ctx context.Context, contractAddress string, rawQueryData []byte, opts ...grpc.CallOption) ([]byte, error)
	QuerySmartContractState(ctx context.Context, contractAddress string, rawQueryData []byte, opts ...grpc.CallOption) ([]byte, error)
}

// QueryClient handles querying contract states
type QueryClient struct {
	conn grpc.ClientConnInterface
}

// NewQueryClient creates a new QueryClient and stores the gRPC connection
func NewQueryClient(conn grpc.ClientConnInterface) *QueryClient {
	return &QueryClient{conn: conn}
}

// QueryRawContractState queries raw contract state
func (q *QueryClient) QueryRawContractState(ctx context.Context, contractAddress string, rawQueryData []byte, opts ...grpc.CallOption) ([]byte, error) {
	in := &wasmtypes.QueryRawContractStateRequest{
		Address:   contractAddress,
		QueryData: rawQueryData,
	}
	out := new(wasmtypes.QueryRawContractStateResponse)

	err := q.conn.Invoke(ctx, "/cosmwasm.wasm.v1.Query/RawContractState", in, out, opts...)
	if err != nil {
		return nil, err
	}

	return out.Data, nil
}

// QuerySmartContractState queries smart contract state
func (q *QueryClient) QuerySmartContractState(ctx context.Context, contractAddress string, rawQueryData []byte, opts ...grpc.CallOption) ([]byte, error) {
	in := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddress,
		QueryData: rawQueryData,
	}
	out := new(wasmtypes.QuerySmartContractStateResponse)

	err := q.conn.Invoke(ctx, "/cosmwasm.wasm.v1.Query/SmartContractState", in, out, opts...)
	if err != nil {
		return nil, err
	}

	return out.Data, nil
}
