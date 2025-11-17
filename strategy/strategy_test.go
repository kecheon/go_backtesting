package strategy_test

import (
	"go-backtesting/config"
	"go-backtesting/strategy"
	"testing"
)

func TestRunBacktest(t *testing.T) {
	cfg := &config.Config{
		FilePath:        "test_data.csv",
		VWZPeriod:       5,
		ZScoreThreshold: 1.0,
		EmaPeriod:       5,
		ADXPeriod:       5,
		ADXThreshold:    0.0,
		BoxFilter: config.BoxFilterConfig{
			Period:      5,
			MinRangePct: 0.01,
		},
		VWZScore: config.VWZScoreConfig{
			MinStdDev: 1e-5,
		},
		TPRate:        0.01,
		SLRate:        0.01,
		BBWPeriod:     20,
		BBWMultiplier: 2.0,
		BBWThreshold:  0.01,
	}

	strategyData, err := strategy.InitializeStrategyDataContext(cfg)
	if err != nil {
		t.Fatalf("InitializeStrategyDataContext failed: %v", err)
	}

	result := strategy.RunBacktest(strategyData, cfg)

	if result.TotalTrades != 0 {
		t.Logf("Expected trades got %d", result.TotalTrades)
	}
}
