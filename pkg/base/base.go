package base

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	providertypes "github.com/margined-protocol/locust-core/pkg/provider"
	"github.com/margined-protocol/locust-core/pkg/types"
	"go.uber.org/zap"
)

type Runner interface {
	Run()
	GetInfo() interface{}
}

type Strategy struct {
	Mut            sync.RWMutex
	Lock           sync.Mutex
	Logger         *zap.Logger
	Running        atomic.Bool
	MainCtx        context.Context
	MainCancel     context.CancelFunc
	Wg             sync.WaitGroup
	PriceProviders map[string]providertypes.Provider
	PreviousPrices map[string]float64
	LastExecution  time.Time
	Cfg            *types.Config
	Runner         Runner
}

func NewBaseStrategy(cfg *types.Config, logger *zap.Logger, runner Runner) *Strategy {
	mainCtx, mainCancel := context.WithCancel(context.Background())

	return &Strategy{
		Logger:         logger,
		Cfg:            cfg,
		MainCtx:        mainCtx,
		MainCancel:     mainCancel,
		PreviousPrices: make(map[string]float64),
		Runner:         runner,
	}
}

func (s *Strategy) IsRunning() bool {
	return s.Running.Load()
}

func (s *Strategy) GetLastExecutionTime() time.Time {
	s.Mut.RLock()
	defer s.Mut.RUnlock()
	return s.LastExecution
}

func (s *Strategy) GetPrices() (map[string]float64, error) {
	prices := make(map[string]float64)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var fetchError error

	for name, provider := range s.PriceProviders {
		wg.Add(1)
		go func(name string, provider providertypes.Provider) {
			defer wg.Done()
			price, err := provider.FetchPrice(s.MainCtx)
			if err != nil {
				s.Logger.Error("Failed to fetch price", zap.String("provider", name), zap.Error(err))
				mu.Lock()
				fetchError = fmt.Errorf("failed to fetch price from provider %s: %w", name, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			prices[name] = price
			mu.Unlock()
		}(name, provider)
	}

	wg.Wait()

	if fetchError != nil {
		s.Logger.Error("Failed to fetch all required prices")
		return nil, fetchError
	}

	if len(prices) != len(s.PriceProviders) {
		err := fmt.Errorf("failed to fetch all required prices")
		s.Logger.Error(err.Error())
		return nil, err
	}

	return prices, nil
}

func (s *Strategy) Start() error {
	if s.Running.Load() {
		return nil
	}
	s.Running.Store(true)
	s.Wg.Add(1)
	go s.Runner.Run()

	return nil
}

func (s *Strategy) Stop() {
	s.Logger.Info("Stopping strategy")
	if !s.Running.Load() {
		return
	}
	s.Running.Store(false)
	s.MainCancel()
}

func (s *Strategy) Info() interface{} {
	return s.Runner.GetInfo()
}
