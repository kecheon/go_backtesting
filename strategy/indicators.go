package strategy

import (
	"go-backtesting/market"
	"math"

	"github.com/markcheno/go-talib"
)

type MarketState string

const (
	ExpandingBullish      MarketState = "ExpandingBullish"
	ExpandingBearish      MarketState = "ExpandingBearish"
	Squeeze               MarketState = "Squeeze"
	Neutral               MarketState = "Neutral"
	InsufficientData      MarketState = "InsufficientData"
	InsufficientATR       MarketState = "InsufficientATR"
	InsufficientBBW       MarketState = "InsufficientBBW"
	InsufficientBBWSeries MarketState = "InsufficientBBWSeries"
)

// TechnicalIndicators holds the values of all technical indicators for a given candle.
type TechnicalIndicators struct {
	BBState  BBWState
	PlusDI   float64
	MinusDI  float64
	VWZScore float64
	ZScore   float64
	EmaShort float64
	EmaLong  float64
	ADX      float64
	DX       float64
}

// StrategyDataContext holds all the data required for a strategy.
type StrategyDataContext struct {
	Candles    market.CandleSticks
	EmaShort   []float64
	EmaLong    []float64
	ZScores    []float64
	VwzScores  []float64
	PlusDI     []float64
	MinusDI    []float64
	AdxSeries  []float64
	BbwzScores []float64
	Bbw        []float64
	DX         []float64
}

// createTechnicalIndicators creates a TechnicalIndicators struct for a given index.
func (s *StrategyDataContext) createTechnicalIndicators(i int) TechnicalIndicators {
	return TechnicalIndicators{
		BBState:  DetectBBWState(s.Candles[:i+1], 20, 2.0, 0),
		PlusDI:   s.PlusDI[i],
		MinusDI:  s.MinusDI[i],
		VWZScore: s.VwzScores[i],
		ZScore:   s.ZScores[i],
		EmaShort: s.EmaShort[i],
		EmaLong:  s.EmaLong[i],
		ADX:      s.AdxSeries[i],
		DX:       s.DX[i],
	}
}

// determineEntrySignal determines the entry signal based on the indicators.
func determineEntrySignal(indicators TechnicalIndicators, adxThreshold float64) (string, bool) {
	longCondition := indicators.EmaShort > indicators.EmaLong &&
		indicators.ZScore < 0.0

	shortCondition := indicators.EmaShort < indicators.EmaLong &&
		indicators.ZScore > 0.0

	if indicators.ADX > adxThreshold &&
		indicators.ADX < 50 &&
		(longCondition || shortCondition) {
		if longCondition {
			return "long", true
		}
		return "short", true
	}

	return "", false
}

func Ema(zscores []float64, period int) []float64 {
	startIdx := 0
	for startIdx < len(zscores) && math.IsNaN(zscores[startIdx]) {
		startIdx++
	}

	if len(zscores[startIdx:]) < period {
		result := make([]float64, len(zscores))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	smaValues := talib.Ema(zscores[startIdx:], period)

	result := make([]float64, len(zscores))
	for i := 0; i < startIdx; i++ {
		result[i] = math.NaN()
	}

	copy(result[startIdx:], smaValues)

	return result
}

func VWZScores(candles market.CandleSticks, period int, minStdDev float64) []float64 {
	if len(candles) < period {
		return nil
	}

	vwz := make([]float64, len(candles))

	for i := range candles {
		if i < period-1 {
			vwz[i] = math.NaN()
			continue
		}

		window := candles[i-period+1 : i+1]

		var weightedSum, weightSum float64
		for _, c := range window {
			weightedSum += c.Close * c.Vol
			weightSum += c.Vol
		}
		if weightSum == 0 {
			vwz[i] = math.NaN()
			continue
		}
		mean := weightedSum / weightSum

		var variance float64
		for _, c := range window {
			diff := c.Close - mean
			variance += c.Vol * diff * diff
		}
		std := math.Sqrt(variance / weightSum)

		if std < minStdDev {
			vwz[i] = math.NaN()
		} else {
			vwz[i] = (candles[i].Close - mean) / std
		}
	}

	return vwz
}

func ZScores(candles market.CandleSticks, period int) []float64 {
	if len(candles) < period {
		return nil
	}

	data := make([]float64, 0, len(candles))
	for _, c := range candles {
		data = append(data, c.Close)
	}

	mean := talib.Ma(data, period, talib.SMA)

	std := talib.StdDev(data, period, 1.0)

	zscores := make([]float64, len(data))
	for i := range data {
		if math.IsNaN(mean[i]) || math.IsNaN(std[i]) || std[i] == 0 {
			zscores[i] = math.NaN()
		} else {
			zscores[i] = (data[i] - mean[i]) / std[i]
		}
	}

	return zscores
}

func BoxFilter(candles market.CandleSticks, period int, minRangePct float64) []bool {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))

	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
	}
	highestHighs := talib.Max(highs, period)
	lowestLows := talib.Min(lows, period)

	isRanging := make([]bool, len(highestHighs))
	for i := range isRanging {
		isRanging[i] = ((highestHighs[i] - lowestLows[i]) / lowestLows[i]) < minRangePct
	}
	return isRanging
}
