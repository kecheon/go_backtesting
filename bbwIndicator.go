package main

import (
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
	candles CandleSticks,
	period int,
	multiplier float64,
	bbwThreshold float64,
) BBWState {
	if len(candles) < period*2 {
		// fmt.Println("+++++++++++++++++++++++=")
		// fmt.Println("insufficuent data case1")
		return BBWState{Status: "InsufficientData"}
	}

	bbwSeries, _, middle, _ := BBW(candles, period, multiplier)
	normalizedBBW := NormalizeBBW(bbwSeries, period)

	if math.Abs(normalizedBBW[len(normalizedBBW)-1]) < bbwThreshold {
		return BBWState{Status: Neutral}
	}

	// === ATR 계산 ===
	atr := ATR(candles, 14)
	if len(atr) < 3 {
		return BBWState{Status: "InsufficientATR"}
	}
	atrUp := atr[len(atr)-1] > atr[len(atr)-2] && atr[len(atr)-2] > atr[len(atr)-3]
	// fmt.Printf("atrUp: %v atr0: %f atr1: %f atr2: %f\n", atrUp, atr[len(atr)-1], atr[len(atr)-2], atr[len(atr)-3])
	// === 이동평균 기반 BBW 평균 계산 ===

	bbw := bbwSeries[len(bbwSeries)-1]
	bbwEMA := talib.Ema(bbwSeries, 5)
	if len(bbwEMA) < period {
		// fmt.Println("+++++++++++++++++++++++=")
		// fmt.Println("insufficuent data case1")
		return BBWState{Status: "InsufficientBBWSeries"}
	}
	bbwAvg := bbwEMA[len(bbwEMA)-1]
	bbwPrev := bbwEMA[len(bbwEMA)-2]
	bbwPPrev := bbwEMA[len(bbwEMA)-3]
	// bbwUp := bbw > 1.01*bbwPrev && bbwPrev > 1.01*bbwPPrev
	bbwUp := bbwAvg > bbwPrev && bbwPrev > bbwPPrev && bbw > bbwAvg
	bbwDown := bbwAvg < bbwPrev && bbwPrev < bbwPPrev && bbw < bbwAvg

	close := candles[len(candles)-1].Close
	// up := upper[len(upper)-1]
	// low := lower[len(lower)-1]

	var status MarketState
	var sidewaysDuration int

	if bbwDown {
		status = Squeeze
	} else if bbwUp {
		// status = BreakUp
		if close > middle[len(middle)-1] && middle[len(middle)-1] > middle[len(middle)-2] {
			// 종가가  중심선 위에 있고 중심선이 상승
			if atrUp {
				// fmt.Println("+++++++++++++++++++++++=")
				// fmt.Println("ExpandingBullish case1")
				status = ExpandingBullish
			} else {
				status = Neutral
			}
		} else if close < middle[len(middle)-1] && middle[len(middle)-1] < middle[len(middle)-2] {
			// 종가가  중심선 아래에 있고 중심선이 하락
			if atrUp {
				// fmt.Println("+++++++++++++++++++++++=")
				// fmt.Println("ExpandingBearish case1")
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
	/*
		•	노이즈(무시): -0.5 <= BBW_z <= +0.5
			— 이 구간은 “평균 ±0.5σ”로, 대부분의 작은 진동(노이즈)을 포함합니다.
		•	약한 신호(주의): BBW_z <= -1.0 → 수축(매우 낮음), BBW_z >= +1.0 → 확장(매우 높음)
			— 표준적 의미의 1σ 이상: 유의미한 변화로 간주.
		•	강한 신호(확인 필요): BBW_z <= -1.5 또는 BBW_z >= +1.5
			— 거의 드문 사건, 더 강력한 신호(추세 전개·볼래틸리티 폭발 가능).
	*/
	if len(bbwSeries) < window {
		return nil
	}

	normalized := make([]float64, len(bbwSeries))
	for i := range bbwSeries {
		if i < window {
			normalized[i] = 0 // 초기값
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

func BBW(candles CandleSticks, period int, multiplier float64) (bbwSeries, upper, middle, lower []float64) {

	closes := make([]float64, 0, len(candles))
	// highs := make([]float64, 0, len(candles))
	// lows := make([]float64, 0, len(candles))
	for _, c := range candles {
		closes = append(closes, c.Close)
		// highs = append(highs, c.High)
		// lows = append(lows, c.Low)
	}

	// === 볼린저 밴드 계산 ===
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

func ATR(candles CandleSticks, period int) (atr []float64) {
	closes := make([]float64, 0, len(candles))
	highs := make([]float64, 0, len(candles))
	lows := make([]float64, 0, len(candles))
	for _, c := range candles {
		closes = append(closes, c.Close)
		highs = append(highs, c.High)
		lows = append(lows, c.Low)
	}

	// === ATR 계산 ===
	atr = talib.Atr(highs, lows, closes, 14)
	if len(atr) < 3 {
		return []float64{}
	}

	return atr
}
