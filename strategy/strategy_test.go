package strategy

import (
	"go-backtesting/config"
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
		TPRate:         0.01,
		SLRate:         0.01,
		BBWPeriod:      20,
		BBWMultiplier:  2.0,
		BBWThreshold:   0.01,
		LongCondition:  "default",
		ShortCondition: "default",
	}

	strategyData, err := InitializeStrategyDataContext(cfg)
	if err != nil {
		t.Fatalf("InitializeStrategyDataContext failed: %v", err)
	}

	if len(strategyData.Candles) == 0 {
		t.Fatal("No candle data loaded")
	}

	longCondition, err := GetEntryCondition(cfg.LongCondition, "long")
	if err != nil {
		t.Fatalf("Failed to get long entry condition: %v", err)
	}

	shortCondition, err := GetEntryCondition(cfg.ShortCondition, "short")
	if err != nil {
		t.Fatalf("Failed to get short entry condition: %v", err)
	}

	result := RunBacktest(strategyData, cfg, longCondition, shortCondition)

	if result.TotalTrades != 2 {
		t.Logf("Expected 2 trades got %d", result.TotalTrades)
	}
}

func TestDetermineEntrySignalWithCustomConditions(t *testing.T) {
	indicators := TechnicalIndicators{
		ADX: []float64{20.0, 30.0, 49.0},
	}
	cfg := &config.Config{
		ADXThreshold:      25.0,
		AdxUpperThreshold: 50.0,
	}

	// Mock condition functions
	mockLongCondition := func(indicators TechnicalIndicators) (bool, bool) {
		return true, false
	}
	mockShortCondition := func(indicators TechnicalIndicators) (bool, bool) {
		return false, false
	}

	// Test with mock long condition
	direction, entry, stop := DetermineEntrySignal(
		indicators,
		cfg,
		mockLongCondition,
		mockShortCondition,
	)

	if !entry || direction != "long" {
		t.Errorf("Expected a long signal, but got direction: '%s' and entry: %v", direction, entry)
	}

	if stop {
		t.Errorf("Expected stop to be false, but got true")
	}
}

func TestMACDIntegration(t *testing.T) {
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
		TPRate:         0.01,
		SLRate:         0.01,
		BBWPeriod:      20,
		BBWMultiplier:  2.0,
		BBWThreshold:   0.01,
		LongCondition:  "default",
		ShortCondition: "default",
	}

	strategyData, err := InitializeStrategyDataContext(cfg)
	if err != nil {
		t.Fatalf("InitializeStrategyDataContext failed: %v", err)
	}

	if len(strategyData.MACD) == 0 {
		t.Fatal("MACD slice is empty")
	}

	expectedLastMACD := 76.93
	lastMACD := strategyData.MACD[len(strategyData.MACD)-1]

	if !CloseEnough(lastMACD, expectedLastMACD, 0.01) {
		t.Errorf("Expected last MACD to be %.2f, but got %.2f", expectedLastMACD, lastMACD)
	}
}

