package analysis

import (
	"math"
	"go-backtesting/internal/model"
)

// CalculatePattern checks for candlestick patterns.
// It returns a list of detected patterns for the *current* candle (last in slice).
// The patterns are: "Hammer", "Engulfing", "ShootingStar", "Doji"

func IsDoji(c model.Candle) bool {
	body := math.Abs(c.Close - c.Open)
	totalRange := c.High - c.Low
	if totalRange == 0 {
		return true // Flat line is a Doji
	}
	// BodySize < (TotalRange * 0.1)
	return body < (totalRange * 0.1)
}

func IsBullishHammer(c model.Candle) bool {
	body := math.Abs(c.Close - c.Open)
	upperWick := c.High - math.Max(c.Open, c.Close)
	lowerWick := math.Min(c.Open, c.Close) - c.Low

	// (LowerWick > Body * 2) && (UpperWick < Body * 0.5)
	if body == 0 { return false }
	return (lowerWick > body*2) && (upperWick < body*0.5)
}

func IsBearishShootingStar(c model.Candle) bool {
	body := math.Abs(c.Close - c.Open)
	upperWick := c.High - math.Max(c.Open, c.Close)
	lowerWick := math.Min(c.Open, c.Close) - c.Low

	// (UpperWick > Body * 2) && (LowerWick < Body * 0.5)
	if body == 0 { return false }
	return (upperWick > body*2) && (lowerWick < body*0.5)
}

func IsBullishEngulfing(prev, curr model.Candle) bool {
	// (PrevClose < PrevOpen) && (CurrClose > PrevOpen) && (CurrClose > CurrOpen)
	return (prev.Close < prev.Open) && (curr.Close > prev.Open) && (curr.Close > curr.Open)
}

func IsBearishEngulfing(prev, curr model.Candle) bool {
	// (PrevClose > PrevOpen) && (CurrClose < PrevOpen) && (CurrClose < CurrOpen)
	return (prev.Close > prev.Open) && (curr.Close < prev.Open) && (curr.Close < curr.Open)
}

// CheckPatterns returns (isBullish, isBearish, patternName)
func CheckPatterns(candles []model.Candle) (bool, bool, string) {
	if len(candles) < 2 {
		return false, false, ""
	}
	curr := candles[len(candles)-1]
	prev := candles[len(candles)-2]

	isDoji := IsDoji(curr)
	if isDoji {
		return false, false, "Doji"
	}

	// Bullish Checks
	if IsBullishHammer(curr) {
		return true, false, "Bullish Hammer"
	}
	if IsBullishEngulfing(prev, curr) {
		return true, false, "Bullish Engulfing"
	}

	// Bearish Checks
	if IsBearishShootingStar(curr) {
		return false, true, "Bearish Shooting Star"
	}
	if IsBearishEngulfing(prev, curr) {
		return false, true, "Bearish Engulfing"
	}

	return false, false, ""
}
