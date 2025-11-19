package strategy

import (
	"go-backtesting/config"
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
	BBState       BBWState
	PlusDI        []float64
	MinusDI       []float64
	VWZScore      []float64
	ZScore        []float64
	EmaShort      []float64
	EmaLong       []float64
	ADX           []float64
	DX            []float64
	BBW           []float64
	BbwzScore     []float64
	MACD          []float64
	MACDSignal    []float64
	MACDHistogram []float64
}

// StrategyDataContext holds all the data required for a strategy.
type StrategyDataContext struct {
	Candles       market.CandleSticks
	EmaShort      []float64
	EmaLong       []float64
	ZScores       []float64
	VwzScores     []float64
	PlusDI        []float64
	MinusDI       []float64
	AdxSeries     []float64
	BbwzScores    []float64
	Bbw           []float64
	DX            []float64
	MACD          []float64
	MACDSignal    []float64
	MACDHistogram []float64
}

// createTechnicalIndicators creates a TechnicalIndicators struct for a given index,
// populating it with the last 3 values of each indicator.
func (s *StrategyDataContext) createTechnicalIndicators(i int, config *config.Config) TechnicalIndicators {
	return TechnicalIndicators{
		BBState:       DetectBBWState(s.Candles[:i+1], config.BBWPeriod, config.BBWMultiplier, config.BBWThreshold),
		PlusDI:        getLastThree(s.PlusDI, i),
		MinusDI:       getLastThree(s.MinusDI, i),
		VWZScore:      getLastThree(s.VwzScores, i),
		ZScore:        getLastThree(s.ZScores, i),
		EmaShort:      getLastThree(s.EmaShort, i),
		EmaLong:       getLastThree(s.EmaLong, i),
		ADX:           getLastThree(s.AdxSeries, i),
		DX:            getLastThree(s.DX, i),
		BBW:           getLastThree(s.Bbw, i),
		BbwzScore:     getLastThree(s.BbwzScores, i),
		MACD:          getLastThree(s.MACD, i),
		MACDSignal:    getLastThree(s.MACDSignal, i),
		MACDHistogram: getLastThree(s.MACDHistogram, i),
	}
}

// getLastThree safely retrieves the last 1, 2, or 3 values from a slice up to a given index.
func getLastThree(data []float64, index int) []float64 {
	start := index - 2
	if start < 0 {
		start = 0
	}
	// The end index for slicing is exclusive, so `index + 1` includes the element at `index`.
	end := index + 1
	if end > len(data) {
		end = len(data)
	}
	return data[start:end]
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
