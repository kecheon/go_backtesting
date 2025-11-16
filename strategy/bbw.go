package strategy

import (
	"go-backtesting/market"
	"math"

	"github.com/markcheno/go-talib"
	"gonum.org/v1/gonum/stat"
)

type BBWState struct {
	Status       MarketState
	BBW          float64
	BBWAvg       float64
	BBWTrendUp   bool
	DurationSide int
}

func DetectBBWState(
	candles market.CandleSticks,
	period int,
	multiplier float64,
	bbwThreshold float64,
) BBWState {
	if len(candles) < period*2 {
		return BBWState{Status: "InsufficientData"}
	}

	bbwSeries, _, middle, _ := BBW(candles, period, multiplier)
	normalizedBBW := NormalizeBBW(bbwSeries, period)

	if math.Abs(normalizedBBW[len(normalizedBBW)-1]) < bbwThreshold {
		return BBWState{Status: Neutral}
	}

	atr := ATR(candles, 14)
	if len(atr) < 3 {
		return BBWState{Status: "InsufficientATR"}
	}
	atrUp := atr[len(atr)-1] > atr[len(atr)-2] && atr[len(atr)-2] > atr[len(atr)-3]

	bbw := bbwSeries[len(bbwSeries)-1]
	bbwEMA := talib.Ema(bbwSeries, 5)
	if len(bbwEMA) < period {
		return BBWState{Status: "InsufficientBBWSeries"}
	}
	bbwAvg := bbwEMA[len(bbwEMA)-1]
	bbwPrev := bbwEMA[len(bbwEMA)-2]
	bbwPPrev := bbwEMA[len(bbwEMA)-3]
	bbwUp := bbwAvg > bbwPrev && bbwPrev > bbwPPrev && bbw > bbwAvg
	bbwDown := bbwAvg < bbwPrev && bbwPrev < bbwPPrev && bbw < bbwAvg

	close := candles[len(candles)-1].Close

	var status MarketState
	var sidewaysDuration int

	if bbwDown {
		status = Squeeze
	} else if bbwUp {
		if close > middle[len(middle)-1] && middle[len(middle)-1] > middle[len(middle)-2] {
			if atrUp {
				status = ExpandingBullish
			} else {
				status = Neutral
			}
		} else if close < middle[len(middle)-1] && middle[len(middle)-1] < middle[len(middle)-2] {
			if atrUp {
				status = ExpandingBearish
			} else {
				status = Neutral
			}
		}
	} else {
		m1 := middle[len(middle)-1]
		m2 := middle[len(middle)-2]

		isCenterLineUp := close > m1 && m1 > m2
		isCenterLineDown := close < m1 && m1 < m2

		if isCenterLineUp && atrUp {
			status = ExpandingBullish
		} else if isCenterLineDown && atrUp {
			status = ExpandingBearish
		} else {
			status = Neutral
		}
	}

	return BBWState{
		Status:       status,
		BBW:          bbw,
		BBWAvg:       bbwAvg,
		BBWTrendUp:   bbwUp,
		DurationSide: sidewaysDuration,
	}
}

func NormalizeBBW(bbwSeries []float64, window int) []float64 {
	if len(bbwSeries) < window {
		return nil
	}

	normalized := make([]float64, len(bbwSeries))
	for i := range bbwSeries {
		if i < window {
			normalized[i] = 0 // initial value
			continue
		}

		windowData := bbwSeries[i-window : i]
		mean := stat.Mean(windowData, nil)
		std := stat.StdDev(windowData, nil)
		if std == 0 {
			normalized[i] = 0
		} else {
			normalized[i] = (bbwSeries[i] - mean) / std
		}
	}
	return normalized
}

func BBW(candles market.CandleSticks, period int, multiplier float64) (bbwSeries, upper, middle, lower []float64) {
	closes := make([]float64, 0, len(candles))
	for _, c := range candles {
		closes = append(closes, c.Close)
	}

	if period == 0 {
		period = 20
	}
	if multiplier == 0 {
		multiplier = 1.5
	}
	upper, middle, lower = talib.BBands(closes, period, multiplier, multiplier, talib.EMA)
	bbwSeries = make([]float64, len(upper))
	for i := range upper {
		if middle[i] == 0 {
			bbwSeries[i] = 0
		} else {
			bbwSeries[i] = (upper[i] - lower[i]) / middle[i]
		}
	}
	return bbwSeries, upper, middle, lower
}

func ATR(candles market.CandleSticks, period int) (atr []float64) {
	closes := make([]float64, 0, len(candles))
	highs := make([]float64, 0, len(candles))
	lows := make([]float64, 0, len(candles))
	for _, c := range candles {
		closes = append(closes, c.Close)
		highs = append(highs, c.High)
		lows = append(lows, c.Low)
	}

	atr = talib.Atr(highs, lows, closes, 14)
	if len(atr) < 3 {
		return []float64{}
	}

	return atr
}
