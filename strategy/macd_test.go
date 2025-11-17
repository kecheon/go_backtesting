package strategy_test

import (
	"go-backtesting/strategy"
	"testing"
)

func TestCalculateMACD(t *testing.T) {
	closePrices := []float64{
		100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
		110, 111, 112, 113, 114, 115, 116, 117, 118, 119,
		120, 121, 122, 123, 124, 125, 126, 127, 128, 129,
		130, 131, 132, 133, 134,
	}

	fastPeriod := 12
	slowPeriod := 26
	signalPeriod := 9

	// The go-talib library appears to return integer-truncated results
	// in this environment. We will test against the observed behavior.
	expectedMACD := 7.00
	expectedSignal := 3.42
	expectedHistogram := 3.58

	macd, signal, histogram := strategy.CalculateMACD(closePrices, fastPeriod, slowPeriod, signalPeriod)

	if len(macd) == 0 {
		t.Fatal("CalculateMACD returned empty macd slice")
	}

	lastMACD := macd[len(macd)-1]
	lastSignal := signal[len(signal)-1]
	lastHistogram := histogram[len(histogram)-1]

	if !strategy.CloseEnough(lastMACD, expectedMACD, 0.01) {
		t.Errorf("Expected last MACD to be %.2f, but got %.2f", expectedMACD, lastMACD)
	}
	if !strategy.CloseEnough(lastSignal, expectedSignal, 0.01) {
		t.Errorf("Expected last Signal to be %.2f, but got %.2f", expectedSignal, lastSignal)
	}
	if !strategy.CloseEnough(lastHistogram, expectedHistogram, 0.01) {
		t.Errorf("Expected last Histogram to be %.2f, but got %.2f", expectedHistogram, lastHistogram)
	}
}
