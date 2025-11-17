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

	if len(strategyData.Candles) == 0 {
		t.Fatal("No candle data loaded")
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

	if len(strategyData.MACD) == 0 {
		t.Fatal("MACD slice is empty")
	}

	expectedLastMACD := 76.93
	lastMACD := strategyData.MACD[len(strategyData.MACD)-1]

	if !strategy.CloseEnough(lastMACD, expectedLastMACD, 0.01) {
		t.Errorf("Expected last MACD to be %.2f, but got %.2f", expectedLastMACD, lastMACD)
	}
}

func TestMACDLongCondition(t *testing.T) {
	// Bullish crossover: histogram was negative, now positive
	indicators := strategy.TechnicalIndicators{
		PrevMACDHistogram: -0.5,
		MACDHistogram:     0.5,
	}
	if !strategy.MACDLongCondition(indicators) {
		t.Error("Expected MACDLongCondition to be true for a bullish crossover")
	}

	// No crossover
	indicators = strategy.TechnicalIndicators{
		PrevMACDHistogram: 0.5,
		MACDHistogram:     1.0,
	}
	if strategy.MACDLongCondition(indicators) {
		t.Error("Expected MACDLongCondition to be false when histogram is still positive")
	}
}

func TestMACDShortCondition(t *testing.T) {
	// Bearish crossover: histogram was positive, now negative
	indicators := strategy.TechnicalIndicators{
		PrevMACDHistogram: 0.5,
		MACDHistogram:     -0.5,
	}
	if !strategy.MACDShortCondition(indicators) {
		t.Error("Expected MACDShortCondition to be true for a bearish crossover")
	}

	// No crossover
	indicators = strategy.TechnicalIndicators{
		PrevMACDHistogram: -0.5,
		MACDHistogram:     -1.0,
	}
	if strategy.MACDShortCondition(indicators) {
		t.Error("Expected MACDShortCondition to be false when histogram is still negative")
	}
}
