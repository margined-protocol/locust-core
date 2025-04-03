package chains

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ChainInfo represents the info of a chain held in the registry
type ChainInfo struct {
	Name      string `json:"pretty_name"`
	ChainType string `json:"chain_type"`
	ChainID   string `json:"chain_id"`
	Prefix    string `json:"bech32_prefix"`
}

// Chain represents a single chain in the registry
type Chain struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	MTime string `json:"mtime"`
}

// Client handles fetching and querying the asset list
type Client struct {
	url        string
	cache      *[]Chain
	cacheMu    sync.RWMutex
	chainCache map[string]ChainInfo
	chainMu    sync.RWMutex
	stopChan   chan struct{}
}

// NewAssetListClient initializes a new AssetListClient
func NewClient(url string) *Client {
	return &Client{
		url:        url,
		stopChan:   make(chan struct{}),
		chainCache: make(map[string]ChainInfo),
	}
}

// fetchRegistry fetches the asset list from the URL and caches it
func (c *Client) fetchRegistry(l *zap.Logger) error {
	l.Info("Fetching registry.json", zap.String("url", c.url))

	client := &http.Client{}
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		l.Error("Failed to create request", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		l.Error("Failed to fetch registry.json", zap.Error(err))
		return fmt.Errorf("failed to fetch registry.json: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Error("Non-200 status code received", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("Failed to read response body", zap.Error(err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var registry []Chain
	err = json.Unmarshal(body, &registry)
	if err != nil {
		l.Error("Failed to unmarshal JSON", zap.Error(err))
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = &registry

	l.Info("Registry refreshed successfully")
	return nil
}

// fetchPrefixes fetches the prefixes from the URL and caches them
func (c *Client) fetchPrefixes(l *zap.Logger) error {
	c.cacheMu.RLock()
	if c.cache == nil {
		c.cacheMu.RUnlock()
		l.Info("Registry not fetched, running fetchRegistry first")
		if err := c.fetchRegistry(l); err != nil {
			l.Error("Failed to fetch registry before fetching prefixes", zap.Error(err))
			return fmt.Errorf("failed to fetch registry before fetching prefixes: %w", err)
		}
	} else {
		c.cacheMu.RUnlock()
	}

	client := &http.Client{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(*c.cache))

	l.Info("Fetching chain.json")

	for _, chain := range *c.cache {
		wg.Add(1)

		go func(chain Chain) {
			defer wg.Done()

			req, err := http.NewRequest("GET", c.url+"/"+chain.Name+"/chain.json", nil)
			if err != nil {
				errChan <- fmt.Errorf("failed to create request for %s: %w", chain.Name, err)
				return
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

			resp, err := client.Do(req)
			if err != nil {
				l.Error("Failed to fetch chain.json", zap.String("chain", chain.Name), zap.Error(err))
				errChan <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				l.Error("Non-200 status code received", zap.String("chain", chain.Name), zap.Int("status", resp.StatusCode))
				errChan <- fmt.Errorf("non-200 status code for %s: %d", chain.Name, resp.StatusCode)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				l.Error("Failed to read response body", zap.String("chain", chain.Name), zap.Error(err))
				errChan <- err
				return
			}

			var info ChainInfo
			err = json.Unmarshal(body, &info)
			if err != nil {
				l.Error("Failed to unmarshal JSON", zap.String("chain", chain.Name), zap.Error(err))
				errChan <- err
				return
			}

			mu.Lock()
			c.chainCache[info.ChainID] = info
			mu.Unlock()
		}(chain)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			l.Error("Error during fetch", zap.Error(err))
		}
	}

	l.Info("Asset list refreshed successfully")
	return nil
}

// StartBackgroundRefresher starts a background process to refresh the cache at regular intervals
func (c *Client) StartBackgroundRefresher(l *zap.Logger, interval time.Duration) {
	refresh := func() {
		if err := c.fetchRegistry(l); err != nil {
			l.Error("Failed to refresh registry", zap.Error(err))
		}
		if err := c.fetchPrefixes(l); err != nil {
			l.Error("Failed to refresh prefixes", zap.Error(err))
		}
	}

	refresh()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				refresh()
			case <-c.stopChan:
				l.Info("Background refresher stopped")
				return
			}
		}
	}()
}

// StopBackgroundRefresher stops the background refresh process
func (c *Client) StopBackgroundRefresher() {
	close(c.stopChan)
}

// GetChainIDPrefix retrieves the prefix for a given chain
func (c *Client) GetChainIDPrefix(l *zap.Logger, denom string) (string, error) {
	c.chainMu.RLock()
	if c.cache == nil {
		c.chainMu.RUnlock()
		if err := c.fetchPrefixes(l); err != nil {
			l.Error("Failed to fetch prefixes", zap.Error(err))
			return "", fmt.Errorf("failed to fetch prefixes: %w", err)
		}
	} else {
		c.chainMu.RUnlock()
	}

	c.chainMu.RLock()
	defer c.chainMu.RUnlock()

	entry, exists := c.chainCache[denom]
	if !exists {
		l.Error("Denom not found in cache", zap.String("denom", denom))
		return "", fmt.Errorf("denom '%s' not found in cache", denom)
	}

	return entry.Prefix, nil
}
