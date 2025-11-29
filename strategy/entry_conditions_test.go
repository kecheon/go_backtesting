package strategy

import (
	"testing"
)

func TestDefaultLongCondition(t *testing.T) {
	indicators := TechnicalIndicators{
		EmaShort: []float64{10.0, 12.0, 15.0},
		EmaLong:  []float64{8.0, 10.0, 12.0},
		ZScore:   []float64{1.0, -0.5, -1.0},
	}

	entry, stop := DefaultLongCondition(indicators)

	if !entry {
		t.Errorf("Expected entry to be true, but got false")
	}

	if stop {
		t.Errorf("Expected stop to be false, but got true")
	}
}

func TestMACDLongCondition(t *testing.T) {
	// Bullish crossover: histogram was negative, now positive
	indicators := TechnicalIndicators{
		MACDHistogram: []float64{-0.5, 0.5},
	}
	entry, stop := MACDLongCondition(indicators)
	if !entry {
		t.Error("Expected MACDLongCondition to be true for a bullish crossover")
	}
	if stop {
		t.Error("Expected stop to be false")
	}

	// No crossover
	indicators = TechnicalIndicators{
		MACDHistogram: []float64{0.5, 1.0},
	}
	entry, stop = MACDLongCondition(indicators)
	if entry {
		t.Error("Expected MACDLongCondition to be false when histogram is still positive")
	}
	if stop {
		t.Error("Expected stop to be false")
	}
}

func TestMACDShortCondition(t *testing.T) {
	// Bearish crossover: histogram was positive, now negative
	indicators := TechnicalIndicators{
		MACDHistogram: []float64{0.5, -0.5},
	}
	entry, stop := MACDShortCondition(indicators)
	if !entry {
		t.Error("Expected MACDShortCondition to be true for a bearish crossover")
	}
	if stop {
		t.Error("Expected stop to be false")
	}

	// No crossover
	indicators = TechnicalIndicators{
		MACDHistogram: []float64{-0.5, -1.0},
	}
	entry, stop = MACDShortCondition(indicators)
	if entry {
		t.Error("Expected MACDShortCondition to be false when histogram is still negative")
	}
	if stop {
		t.Error("Expected stop to be false")
	}
}
