package perps

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"go.uber.org/zap"
)

// can be used for any provider
// verifyPositionChange checks if the position has changed by the expected amount
func verifyPositionChange(ctx context.Context, l *zap.Logger, m Provider, initialPosition *Position, expectedSizeChange sdkmath.Int) error {
	// Wait a short time for the order to be processed
	time.Sleep(1 * time.Second)

	// Check position multiple times to confirm the change
	for attempts := 0; attempts < 3; attempts++ {
		newPosition, err := m.GetPosition(ctx)
		if err != nil {
			l.Warn("Failed to get updated position",
				zap.Error(err),
				zap.Int("attempt", attempts+1),
			)
			time.Sleep(1 * time.Second)
			continue
		}

		// Calculate actual size change
		actualChange := newPosition.Amount.Sub(initialPosition.Amount)

		// Check if the position changed as expected
		if actualChange.Equal(expectedSizeChange) {
			l.Info("Position change verified",
				zap.String("initial_size", initialPosition.Amount.String()),
				zap.String("final_size", newPosition.Amount.String()),
				zap.String("change", actualChange.String()),
				zap.String("expected_change", expectedSizeChange.String()),
			)
			return nil
		}

		l.Debug("Position not yet updated",
			zap.String("expected_change", expectedSizeChange.String()),
			zap.String("actual_change", actualChange.String()),
			zap.Int("attempt", attempts+1),
		)

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("position change not confirmed: expected change %s, got different value", expectedSizeChange)
}
