package strategy

import (
	"go-backtesting/config"
	"go-backtesting/market"
	"math"
)

// Helper to check if candle touches any level in a list within proximity
func isTouchingLevel(candle market.Candle, levels []float64, proximityPct float64) bool {
	for _, lvl := range levels {
		threshold := lvl * (proximityPct / 100.0)
		levelHigh := lvl + threshold
		levelLow := lvl - threshold

		// Check intersection between Candle Range [Low, High] and Level Zone [Level-Thresh, Level+Thresh]
		// Intersection exists if max(Low, LevelLow) <= min(High, LevelHigh)
		if math.Max(candle.Low, levelLow) <= math.Min(candle.High, levelHigh) {
			return true
		}
	}
	return false
}

// VolumeClusterLongEntry checks for long entry conditions.
// Condition 1: Current Price enters LowerPOC proximity.
// Condition 2: Bullish Hammer or Bullish Engulfing.
// Action: Long Entry.
func VolumeClusterLongEntry(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	if len(indicators.Candles) < 2 {
		return false, false
	}

	// Use the latest completed candle for pattern recognition?
	// Usually strategy runs on the CLOSE of the candle `i`.
	// So `indicators.Candles` ends with the current candle at `i`.
	curr := indicators.Candles[len(indicators.Candles)-1]

	// Check for Doji Filter first
	if IsDoji(curr) {
		return false, false
	}

	// Check Proximity to LowerPOCs (Support)
	// Spec: "Condition 1: current price is within LowerPOC's POC_Proximity"
	proximity := config.VolumeCluster.POCProximity
	if proximity == 0 {
		proximity = 0.2 // default
	}

	isNearSupport := isTouchingLevel(curr, indicators.VolumeProfile.LowerPOCs, proximity)
	// Also check Global POC if price is above it? Or is Global POC always relevant?
	// Spec says "LowerPOC". Global POC can be support too.
	// If price > POC, POC acts as support (Lower).
	// If price < POC, POC acts as resistance (Upper).
	// Let's add logic: if Global POC < Current Price, treat it as potential support too.
	if curr.Close > indicators.VolumeProfile.POC {
		if isTouchingLevel(curr, []float64{indicators.VolumeProfile.POC}, proximity) {
			isNearSupport = true
		}
	}

	if !isNearSupport {
		return false, false
	}

	// Check Patterns
	isHammer := IsBullishHammer(curr)

	isEngulfing := false
	if len(indicators.Candles) >= 2 {
		prev := indicators.Candles[len(indicators.Candles)-2]
		isEngulfing = IsBullishEngulfing(prev, curr)
	}

	if isHammer || isEngulfing {
		return true, false // Entry, No Stop signal here
	}

	// --- Check Exit Signals ---
	// If we are already in a position, we might need to exit based on UpperPOC or Reversal Patterns.
	// Since EntryCondition function signature is (entry, stop), we can return stop=true.
	_, shouldExit := VolumeClusterLongExit(indicators, config)
	if shouldExit {
		return false, true
	}

	return false, false
}

// VolumeClusterLongExit checks for long exit conditions.
// Condition 1: Price reaches UpperPOC.
// Condition 2: Bearish Shooting Star or Bearish Engulfing.
// Target: UpperPOC (Take Profit).
func VolumeClusterLongExit(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	if len(indicators.Candles) < 1 {
		return false, false
	}
	curr := indicators.Candles[len(indicators.Candles)-1]

	// Condition 1: Reach UpperPOC (Resistance)
	// "Target: UpperPOC... 90% Take Profit" - handled by specialized exit logic or here?
	// This function returns (entry, exit). So returning (false, true) triggers exit.

	proximity := config.VolumeCluster.POCProximity
	if proximity == 0 {
		proximity = 0.2
	}

	isNearResistance := isTouchingLevel(curr, indicators.VolumeProfile.UpperPOCs, proximity)

	// Check Global POC as resistance if price < POC
	if curr.Close < indicators.VolumeProfile.POC {
		if isTouchingLevel(curr, []float64{indicators.VolumeProfile.POC}, proximity) {
			isNearResistance = true
		}
	}

	if isNearResistance {
		return false, true
	}

	// Condition 2: Bearish Patterns
	isShootingStar := IsBearishShootingStar(curr)

	isEngulfing := false
	if len(indicators.Candles) >= 2 {
		prev := indicators.Candles[len(indicators.Candles)-2]
		isEngulfing = IsBearishEngulfing(prev, curr)
	}

	if isShootingStar || isEngulfing {
		return false, true
	}

	return false, false
}

// VolumeClusterShortEntry
// Symmetrical to Long Entry.
// Condition 1: Near UpperPOC (Resistance).
// Condition 2: Bearish Pattern.
func VolumeClusterShortEntry(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	if len(indicators.Candles) < 2 {
		return false, false
	}
	curr := indicators.Candles[len(indicators.Candles)-1]

	if IsDoji(curr) {
		return false, false
	}

	proximity := config.VolumeCluster.POCProximity
	if proximity == 0 {
		proximity = 0.2
	}

	isNearResistance := isTouchingLevel(curr, indicators.VolumeProfile.UpperPOCs, proximity)

	if curr.Close < indicators.VolumeProfile.POC {
		if isTouchingLevel(curr, []float64{indicators.VolumeProfile.POC}, proximity) {
			isNearResistance = true
		}
	}

	if !isNearResistance {
		return false, false
	}

	isShootingStar := IsBearishShootingStar(curr)

	isEngulfing := false
	if len(indicators.Candles) >= 2 {
		prev := indicators.Candles[len(indicators.Candles)-2]
		isEngulfing = IsBearishEngulfing(prev, curr)
	}

	if isShootingStar || isEngulfing {
		return true, false
	}

	// --- Check Exit Signals ---
	_, shouldExit := VolumeClusterShortExit(indicators, config)
	if shouldExit {
		return false, true
	}

	return false, false
}

// VolumeClusterShortExit
// Condition 1: Near LowerPOC (Support).
// Condition 2: Bullish Pattern.
func VolumeClusterShortExit(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	if len(indicators.Candles) < 1 {
		return false, false
	}
	curr := indicators.Candles[len(indicators.Candles)-1]

	proximity := config.VolumeCluster.POCProximity
	if proximity == 0 {
		proximity = 0.2
	}

	isNearSupport := isTouchingLevel(curr, indicators.VolumeProfile.LowerPOCs, proximity)

	if curr.Close > indicators.VolumeProfile.POC {
		if isTouchingLevel(curr, []float64{indicators.VolumeProfile.POC}, proximity) {
			isNearSupport = true
		}
	}

	if isNearSupport {
		return false, true
	}

	isHammer := IsBullishHammer(curr)

	isEngulfing := false
	if len(indicators.Candles) >= 2 {
		prev := indicators.Candles[len(indicators.Candles)-2]
		isEngulfing = IsBullishEngulfing(prev, curr)
	}

	if isHammer || isEngulfing {
		return false, true
	}

	return false, false
}
