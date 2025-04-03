package backoff

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

var globalBackoffOptions = []backoff.ExponentialBackOffOpts{
	func(b *backoff.ExponentialBackOff) {
		b.InitialInterval = 1 * time.Second
	},
	func(b *backoff.ExponentialBackOff) {
		b.MaxInterval = 32 * time.Second
	},
	func(b *backoff.ExponentialBackOff) {
		b.Multiplier = 2
	},
	func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 60 * time.Second
	},
}

var lightningBackoffOptions = []backoff.ExponentialBackOffOpts{
	func(b *backoff.ExponentialBackOff) {
		b.InitialInterval = 1 * time.Second
	},
	func(b *backoff.ExponentialBackOff) {
		b.MaxElapsedTime = 2 * time.Second
	},
}

func NewBackoff(_ context.Context) *backoff.ExponentialBackOff {
	return backoff.NewExponentialBackOff(globalBackoffOptions...)
}

func NewLightningBackoff(_ context.Context) *backoff.ExponentialBackOff {
	return backoff.NewExponentialBackOff(lightningBackoffOptions...)
}
