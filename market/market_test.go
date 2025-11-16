package market_test

import (
	"go-backtesting/market"
	"os"
	"testing"
	"time"
)

func TestReadCandlesFromCSV(t *testing.T) {
	// Create a temporary test csv file
	content := `Time,Open,High,Low,Close,Volume
2023-01-01 00:00:00,100,105,95,102,1000
2023-01-01 00:05:00,102,106,101,105,1200
2023-01-01 00:10:00,105,110,104,108,1500`
	tmpfile, err := os.CreateTemp("", "test_data.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test loading the candles
	candles, err := market.ReadCandlesFromCSV(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadCandlesFromCSV failed: %v", err)
	}

	// Verify the loaded candles
	if len(candles) != 3 {
		t.Fatalf("Expected 3 candles, but got %d", len(candles))
	}

	expectedTimes := []time.Time{
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 1, 1, 0, 5, 0, 0, time.UTC),
		time.Date(2023, 1, 1, 0, 10, 0, 0, time.UTC),
	}
	expectedOpens := []float64{100, 102, 105}
	expectedHighs := []float64{105, 106, 110}
	expectedLows := []float64{95, 101, 104}
	expectedCloses := []float64{102, 105, 108}
	expectedVols := []float64{1000, 1200, 1500}

	for i, candle := range candles {
		if !candle.Time.Equal(expectedTimes[i]) {
			t.Errorf("Expected Time to be %v, but got %v", expectedTimes[i], candle.Time)
		}
		if candle.Open != expectedOpens[i] {
			t.Errorf("Expected Open to be %f, but got %f", expectedOpens[i], candle.Open)
		}
		if candle.High != expectedHighs[i] {
			t.Errorf("Expected High to be %f, but got %f", expectedHighs[i], candle.High)
		}
		if candle.Low != expectedLows[i] {
			t.Errorf("Expected Low to be %f, but got %f", expectedLows[i], candle.Low)
		}
		if candle.Close != expectedCloses[i] {
			t.Errorf("Expected Close to be %f, but got %f", expectedCloses[i], candle.Close)
		}
		if candle.Vol != expectedVols[i] {
			t.Errorf("Expected Vol to be %f, but got %f", expectedVols[i], candle.Vol)
		}
	}
}
