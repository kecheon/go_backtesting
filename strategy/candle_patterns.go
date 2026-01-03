package strategy

import (
	"go-backtesting/market"
	"math"
)

// Bullish Hammer: (LowerWick > Body * 2) && (UpperWick < Body * 0.5)
// Bullish body (Close > Open) or Bearish body (Close < Open)?
// Usually Hammer is bullish signal regardless of color, but body is small.
// The spec doesn't specify color, but typical Hammer is small body at top of range.
// "Bullish Hammer" usually implies it appears in downtrend.
// Formula: LowerWick > Body * 2 ...
func IsBullishHammer(c market.Candle) bool {
	body := math.Abs(c.Close - c.Open)
	lowerWick := math.Min(c.Close, c.Open) - c.Low
	upperWick := c.High - math.Max(c.Close, c.Open)

	// Avoid division by zero if body is tiny, treat body as non-zero or use multiplication form
	return lowerWick > body*2.0 && upperWick < body*0.5
}

// Bullish Engulfing: (PrevClose < PrevOpen) && (CurrClose > PrevOpen) && (CurrClose > CurrOpen)
// Prev candle was Bearish, Current candle is Bullish, and Current Close > Prev Open.
// Spec: (PrevClose < PrevOpen) && (CurrClose > PrevOpen) && (CurrClose > CurrOpen)
// Note: Standard Engulfing also requires CurrOpen < PrevClose (gapped down) or similar engulfing of body.
// But we stick to the provided spec strict formula.
func IsBullishEngulfing(prev, curr market.Candle) bool {
	prevBearish := prev.Close < prev.Open
	currBullish := curr.Close > curr.Open // derived from CurrClose > CurrOpen in spec

	return prevBearish &&
	       curr.Close > prev.Open &&
	       currBullish
}

// Bearish Shooting Star: (UpperWick > Body * 2) && (LowerWick < Body * 0.5)
func IsBearishShootingStar(c market.Candle) bool {
	body := math.Abs(c.Close - c.Open)
	lowerWick := math.Min(c.Close, c.Open) - c.Low
	upperWick := c.High - math.Max(c.Close, c.Open)

	return upperWick > body*2.0 && lowerWick < body*0.5
}

// Bearish Engulfing: (PrevClose > PrevOpen) && (CurrClose < PrevOpen) && (CurrClose < CurrOpen)
// Prev candle Bullish, Current candle Bearish, Current Close < Prev Open.
func IsBearishEngulfing(prev, curr market.Candle) bool {
	prevBullish := prev.Close > prev.Open
	currBearish := curr.Close < curr.Open // derived from CurrClose < CurrOpen

	return prevBullish &&
	       curr.Close < prev.Open &&
	       currBearish
}

// Doji: BodySize < (TotalRange * 0.1)
func IsDoji(c market.Candle) bool {
	body := math.Abs(c.Close - c.Open)
	totalRange := c.High - c.Low

	if totalRange == 0 {
		return true
	}

	return body < (totalRange * 0.1)
}
