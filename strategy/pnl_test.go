package strategy

import (
	"testing"
	"time"
)

func TestPartialCloseLongPosition(t *testing.T) {
	initialTrade := &Trade{
		EntryTime:  time.Now(),
		EntryPrice: 100.0,
		Direction:  "long",
		Size:       2.0,
	}
	closePrice := 110.0
	closeSize := 1.0

	realizedPnl := calculatePartialClosePnl(initialTrade, closePrice, closeSize)

	expectedPnl := 10.0 // (110 - 100) * 1.0
	if realizedPnl != expectedPnl {
		t.Errorf("Expected realized PnL to be %.2f, but got %.2f", expectedPnl, realizedPnl)
	}

	if initialTrade.Size != 1.0 {
		t.Errorf("Expected trade size to be updated to 1.0, but got %.2f", initialTrade.Size)
	}
}

func TestAverageIntoLongPosition(t *testing.T) {
	initialTrade := &Trade{
		EntryTime:  time.Now(),
		EntryPrice: 100.0,
		Direction:  "long",
		Size:       1.0,
	}
	addPrice := 110.0
	addSize := 1.0

	calculateAveragePrice(initialTrade, addPrice, addSize)

	expectedPrice := 105.0 // ((1.0 * 100.0) + (1.0 * 110.0)) / 2.0
	if initialTrade.EntryPrice != expectedPrice {
		t.Errorf("Expected new entry price to be %.2f, but got %.2f", expectedPrice, initialTrade.EntryPrice)
	}

	if initialTrade.Size != 2.0 {
		t.Errorf("Expected trade size to be updated to 2.0, but got %.2f", initialTrade.Size)
	}
}
