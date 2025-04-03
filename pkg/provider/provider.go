package types

import (
	"golang.org/x/net/context"
)

// Provider defines an interface a data provider must implement.
type Provider interface {
	// Name returns the name of the provider.
	Name() string

	// FetchPrice fetches the price for a given key.
	FetchPrice(ctx context.Context) (float64, error)
}
