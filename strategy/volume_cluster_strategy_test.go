package strategy

import (
	"go-backtesting/config"
	"go-backtesting/market"
	"testing"
)

func TestVolumeClusterLongEntry(t *testing.T) {
	// Setup
	cfg := &config.Config{
		VolumeCluster: config.VolumeProfileConfig{
			POCProximity: 1.0, // 1%
		},
	}

	// 1. Scenario: Hammer near LowerPOC -> Entry
	lowerPOCs := []float64{99.0} // Support at 99
	// Candle: Open 100, Close 100, Low 99.5, High 100.5 -> Small body (0), LowerWick (0.5).
	// Hammer: LowerWick > Body*2. 0.5 > 0*2.

	// Create a Hammer candle near 99.0
	// Open=99.8, Close=99.9. Body=0.1. LowerWick needs > 0.2.
	// Low=99.5. LowerWick = 99.8-99.5 = 0.3. 0.3 > 0.2 OK.
	// High=99.95. UpperWick = 0.05. UpperWick < 0.05 OK (0.5*Body).

	hammerCandle := market.Candle{
		Open: 99.8, Close: 99.9, High: 99.95, Low: 99.5,
	}

	indicators := TechnicalIndicators{
		VolumeProfile: VolumeProfile{
			LowerPOCs: lowerPOCs,
			POC: 105.0, // Global POC far away
		},
		Candles: []market.Candle{
			{Close: 102.0}, // Prev
			hammerCandle,   // Curr
		},
	}

	entry, stop := VolumeClusterLongEntry(indicators, cfg)
	if !entry {
		t.Errorf("Expected Entry=true for Hammer near LowerPOC")
	}
	if stop {
		t.Errorf("Expected Stop=false")
	}

	// 2. Scenario: Not near POC -> No Entry
	indicators.VolumeProfile.LowerPOCs = []float64{90.0} // Far support
	entry, _ = VolumeClusterLongEntry(indicators, cfg)
	if entry {
		t.Errorf("Expected Entry=false when not near POC")
	}
}

func TestVolumeClusterLongExit(t *testing.T) {
	// Setup
	cfg := &config.Config{
		VolumeCluster: config.VolumeProfileConfig{
			POCProximity: 1.0, // 1%
		},
	}

	// 1. Scenario: Shooting Star near UpperPOC -> Exit
	upperPOCs := []float64{105.0}

	// Shooting Star: UpperWick > Body*2.
	// Open=104.5, Close=104.4. Body=0.1.
	// High=105.0. UpperWick=0.5. 0.5 > 0.2 OK.
	// Low=104.35. LowerWick=0.05.

	starCandle := market.Candle{
		Open: 104.5, Close: 104.4, High: 105.0, Low: 104.35,
	}

	indicators := TechnicalIndicators{
		VolumeProfile: VolumeProfile{
			UpperPOCs: upperPOCs,
			POC: 95.0,
		},
		Candles: []market.Candle{
			{Close: 104.0},
			starCandle,
		},
	}

	_, exit := VolumeClusterLongExit(indicators, cfg)
	if !exit {
		t.Errorf("Expected Exit=true for Shooting Star near UpperPOC")
	}
}
