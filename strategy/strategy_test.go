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

	result := strategy.RunBacktest(strategyData, cfg, strategy.DefaultLongCondition, strategy.DefaultShortCondition)

	if result.TotalTrades != 0 {
		t.Logf("Expected trades got %d", result.TotalTrades)
	}
}

func TestDetermineEntrySignalWithCustomConditions(t *testing.T) {
	indicators := strategy.TechnicalIndicators{
		ADX: 49.0,
	}
	adxThreshold := 25.0

	// Mock condition functions
	mockLongCondition := func(indicators strategy.TechnicalIndicators) bool {
		return true
	}
	mockShortCondition := func(indicators strategy.TechnicalIndicators) bool {
		return false
	}

	// Test with mock long condition
	direction, hasSignal := strategy.DetermineEntrySignal(
		indicators,
		adxThreshold,
		mockLongCondition,
		mockShortCondition,
	)

	if !hasSignal || direction != "long" {
		t.Errorf("Expected a long signal, but got direction: '%s' and hasSignal: %v", direction, hasSignal)
	}
}
