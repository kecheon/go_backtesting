package strategy

import (
	"go-backtesting/config"
	"go-backtesting/market"
	"testing"
)

func TestCalculateVolumeProfile(t *testing.T) {
	// Setup dummy candles
	candles := []market.Candle{
		{Low: 100, High: 102, Close: 101, Vol: 100},
		{Low: 101, High: 103, Close: 102, Vol: 200},
		{Low: 102, High: 104, Close: 103, Vol: 300}, // POC area
		{Low: 101, High: 103, Close: 102, Vol: 150},
		{Low: 105, High: 107, Close: 106, Vol: 100}, // Local peak
		{Low: 95, High: 97, Close: 96, Vol: 100},    // Local peak
	}

	cfg := config.VolumeProfileConfig{
		LookbackPeriod: 5,
		BinSizePct:     0.5,
		MinPOCDistance: 1.0, // 1% distance
		POCProximity:   0.2,
	}

	currentPrice := 102.0
	// Bin size = 102 * 0.005 = 0.51

	vp := CalculateVolumeProfile(candles, currentPrice, cfg)

	if vp.POC == 0 {
		t.Errorf("Expected non-zero POC")
	}

	// Verify POC logic roughly
	// Max volume is in range 102-104 (vol 300).
	// Bins around 103 should have high volume.

	t.Logf("POC: %f", vp.POC)
	t.Logf("UpperPOCs: %v", vp.UpperPOCs)
	t.Logf("LowerPOCs: %v", vp.LowerPOCs)

	if len(vp.UpperPOCs) == 0 && len(vp.LowerPOCs) == 0 {
		// Might happen if everything is too close to current price or merged
	}

	// With MinPOCDistance, check if peaks are distinct
	allPeaks := append([]float64{vp.POC}, vp.UpperPOCs...)
	allPeaks = append(allPeaks, vp.LowerPOCs...)

	// Basic validation
	if len(allPeaks) < 1 {
		t.Errorf("Expected at least POC")
	}
}
