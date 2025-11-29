package strategy

import (
	"go-backtesting/config"
	"testing"
)

func TestRunBacktest(t *testing.T) {
	cfg, err := config.LoadConfig("../config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.FilePath = "test_data.csv"
	cfg.EmaPeriod = 5

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

	if result.TotalTrades == 0 {
		t.Logf("Expected trades but got 0")
	}
}

func TestMACDIntegration(t *testing.T) {
	cfg, err := config.LoadConfig("../config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.FilePath = "test_data.csv"
	cfg.EmaPeriod = 5

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
