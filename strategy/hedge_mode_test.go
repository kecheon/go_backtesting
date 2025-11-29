package strategy

import (
	"go-backtesting/config"
	"go-backtesting/market"
	"testing"
	"time"
)

func TestEnterHedgeModeFromLongPosition(t *testing.T) {
	cfg := &config.Config{
		HedgeMode:           true,
		HedgeSizeMultiplier: 2.0,
		MinPositionSize:     0.1,
	}

	// Initial long trade
	activeLongTrade := &Trade{
		EntryTime:  time.Now(),
		EntryPrice: 100.0,
		Direction:  "long",
		Size:       1.0,
	}
	var activeShortTrade *Trade

	// Create a dummy candle
	currentCandle := market.Candle{Close: 105.0}

	// Run the logic that should create a hedge position
	_, activeShortTrade = handleStopSignal(cfg, activeLongTrade, activeShortTrade, "long", currentCandle)

	if activeShortTrade == nil {
		t.Fatal("Expected a new short trade to be created, but it was nil")
	}

	if activeShortTrade.Direction != "short" {
		t.Errorf("Expected the new trade to be short, but it was %s", activeShortTrade.Direction)
	}

	expectedSize := activeLongTrade.Size * cfg.HedgeSizeMultiplier
	if activeShortTrade.Size != expectedSize {
		t.Errorf("Expected the new short trade to have size %.2f, but it was %.2f", expectedSize, activeShortTrade.Size)
	}
}
