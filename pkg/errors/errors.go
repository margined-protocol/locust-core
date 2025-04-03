package errors

import (
	"errors"
)

var (
	ErrZeroBalance            = errors.New("balance cannot be zero")
	ErrBalanceBelowMinBuySize = errors.New("balance is below the min buy size")
	ErrArbBelowThreshold      = errors.New("arb is below threshold")
	ErrNoValidGranters        = errors.New("no valid granters found")
	ErrFailedToFetchGranters  = errors.New("failed to fetch granters")
	ErrTokensNotInOrder       = errors.New("tokens are not in lexicographical order")
	ErrNoFunds                = errors.New("no withdrawable balance on fund contract")
)
