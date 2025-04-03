package skipgo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
)

// skipGoClient is the concrete implementation of SkipGoClient.
type skipGoClient struct {
	baseURL *url.URL
	http    *http.Client
}

// NewClient creates a new instance of SkipGoClient.
func NewClient(baseURL string) (Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL %s: %w", baseURL, err)
	}

	return &skipGoClient{
		baseURL: parsedURL,
		http:    http.DefaultClient,
	}, nil
}

// Balance fetches the balances for the specified chains.
func (s *skipGoClient) Balance(ctx context.Context, request *BalancesRequest) (*BalancesResponse, error) {
	const endpoint = "/v2/info/balances"
	u, err := s.baseURL.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("joining base URL with endpoint %s: %w", endpoint, err)
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleError(resp.Body)
	}

	var res BalancesResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}

	return &res, nil
}

// Route determines the best route for a given asset transfer.
func (s *skipGoClient) Route(ctx context.Context, sourceAssetDenom, sourceAssetChainID, destAssetDenom, destAssetChainID string, amountIn *big.Int) (*RouteResponse, error) {
	return s.route(ctx, sourceAssetDenom, sourceAssetChainID, destAssetDenom, destAssetChainID, amountIn)
}

// SwapRoute determines the best route for a given asset transfer, assuming source == destination.
func (s *skipGoClient) SwapRoute(ctx context.Context, tokenIn, tokenOut, chainID string, amountIn *big.Int) (*RouteResponse, error) {
	return s.route(ctx, tokenIn, chainID, tokenOut, chainID, amountIn)
}

func (s *skipGoClient) route(ctx context.Context, sourceAssetDenom, sourceAssetChainID, destAssetDenom, destAssetChainID string, amountIn *big.Int) (*RouteResponse, error) {
	const endpoint = "/v2/fungible/route"
	u, err := s.baseURL.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("joining base URL with endpoint %s: %w", endpoint, err)
	}

	body := RouteRequest{
		SourceAssetDenom:   sourceAssetDenom,
		SourceAssetChainID: sourceAssetChainID,
		DestAssetDenom:     destAssetDenom,
		DestAssetChainID:   destAssetChainID,
		AmountIn:           amountIn.String(),
		AllowMultiTx:       true,
		Bridges:            []string{"CCTP", "IBC"},
		AllowUnsafe:        false,
		SmartSwapOptions:   SmartSwapOptions{EVMSwaps: false, SplitRoutes: true},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling route request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleError(resp.Body)
	}

	var res RouteResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding route response: %w", err)
	}

	return &res, nil
}

// Msgs fetches the necessary messages for a transaction.
func (s *skipGoClient) Msgs(
	ctx context.Context,
	route RouteResponse,
	addressList []string,
	slippage string,
) ([]Tx, error) {
	const endpoint = "/v2/fungible/msgs"
	u, err := s.baseURL.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("joining base URL with endpoint %s: %w", endpoint, err)
	}

	body := MsgsRequest{
		SourceAssetDenom:         route.SourceAssetDenom,
		SourceAssetChainID:       route.SourceAssetChainID,
		DestAssetDenom:           route.DestAssetDenom,
		DestAssetChainID:         route.DestAssetChainID,
		AmountIn:                 route.AmountIn,
		AmountOut:                route.AmountOut,
		SlippageTolerancePercent: slippage,
		AddressList:              addressList,
		Operations:               route.Operations,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling msgs request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleError(resp.Body)
	}

	var res MsgsResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding msgs response: %w", err)
	}

	return res.Txs, nil
}

// SubmitTx submits a transaction to the specified chain.
func (s *skipGoClient) SubmitTx(ctx context.Context, tx []byte, chainID string) (TxHash, error) {
	const endpoint = "/v2/tx/submit"
	u, err := s.baseURL.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("joining base URL with endpoint %s: %w", endpoint, err)
	}

	encodedTx := base64.StdEncoding.EncodeToString(tx)
	body := SubmitRequest{
		Tx:      encodedTx,
		ChainID: chainID,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling submit request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating HTTP request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := handleError(resp.Body)
		return "", fmt.Errorf("status code %d returned from SkipGo when submitting transaction: %w", resp.StatusCode, errMsg)
	}

	var res SubmitResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decoding submit response: %w", err)
	}

	return TxHash(res.TxHash), nil
}

// TrackTx tracks the status of a transaction.
func (s *skipGoClient) TrackTx(ctx context.Context, txHash, chainID string) (TxHash, error) {
	const endpoint = "/v2/tx/track"
	u, err := s.baseURL.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("joining base URL with endpoint %s: %w", endpoint, err)
	}

	body := TrackRequest{
		TxHash:  txHash,
		ChainID: chainID,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling track request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating HTTP request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := handleError(resp.Body)
		return "", fmt.Errorf("status code %d returned from SkipGo when tracking transaction: %w", resp.StatusCode, errMsg)
	}

	var res TrackResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decoding track response: %w", err)
	}

	return TxHash(res.TxHash), nil
}

// Status retrieves the status of a transaction.
func (s *skipGoClient) Status(ctx context.Context, tx TxHash, chainID string) (*StatusResponse, error) {
	const endpoint = "/v2/tx/status"

	u, err := s.baseURL.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("joining base URL with endpoint %s: %w", endpoint, err)
	}

	query := u.Query()
	query.Set("tx_hash", string(tx))
	query.Set("chain_id", chainID)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleError(resp.Body)
	}

	var res StatusResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding status response: %w", err)
	}

	return &res, nil
}

// handleError processes the error response from SkipGo.
func handleError(body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("reading error response body: %w", err)
	}

	var e Error
	if err := json.Unmarshal(data, &e); err != nil {
		return fmt.Errorf("decoding SkipGo error: %w", err)
	}

	return e
}
