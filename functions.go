package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/markcheno/go-talib"
)

// domain.Candle and domain.CandleSticks definitions
// Assuming 'domain' is a local package, defining them here for a self-contained script.
type Candle struct {
	Time  time.Time
	Open  float64
	High  float64
	Low   float64
	Close float64
	Vol   float64
}

type CandleSticks []Candle

type EntrySignal struct {
	Time      time.Time
	Price     float64
	Direction string // "long" or "short"
}

// readCandlesFromCSV는 주어진 경로의 CSV 파일을 읽어 CandleSticks로 변환합니다.
func readCandlesFromCSV(filePath string) (CandleSticks, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	_, err = reader.Read() // Skip header
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("error: file is empty or contains only a header")
		}
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	var candles CandleSticks
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading record: %w", err)
		}

		// CSV 포맷: YYYY-MM-DD HH:MM:SS, open, high, low, close, volume
		t, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			log.Printf("Error parsing timestamp, skipping record: %v", err)
			continue
		}

		open, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Printf("Error parsing open price, skipping record: %v", err)
			continue
		}
		high, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Error parsing high price, skipping record: %v", err)
			continue
		}
		low, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Error parsing low price, skipping record: %v", err)
			continue
		}
		close, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			log.Printf("Error parsing close price, skipping record: %v", err)
			continue
		}
		vol, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Printf("Error parsing volume, skipping record: %v", err)
			continue
		}

		candles = append(candles, Candle{
			Time:  t,
			Open:  open,
			High:  high,
			Low:   low,
			Close: close,
			Vol:   vol,
		})
	}
	return candles, nil
}

// EmaVWZScores calculates the EMA of given z-scores using go-talib.
// zscores: input slice of z-score values
// period: EMA period (e.g., 14)
// returns: slice of EMA-smoothed z-scores
func Ema(zscores []float64, period int) []float64 {
	// 1. zscores 슬라이스에서 NaN이 아닌 첫 번째 값의 인덱스를 찾습니다.
	startIdx := 0
	for startIdx < len(zscores) && math.IsNaN(zscores[startIdx]) {
		startIdx++
	}

	// 2. 유효한 데이터 포인트가 SMA를 계산하기에 충분하지 않으면,
	//    NaN으로 채워진 슬라이스를 반환합니다.
	if len(zscores[startIdx:]) < period {
		result := make([]float64, len(zscores))
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// 3. zscores의 유효한 부분에 대해서만 SMA를 계산합니다.
	smaValues := talib.Ema(zscores[startIdx:], period)

	// 4. 최종 결과 슬라이스를 생성하고, 원본 데이터와 길이를 맞추기 위해
	//    시작 부분을 NaN으로 채웁니다.
	result := make([]float64, len(zscores))
	for i := 0; i < startIdx; i++ {
		result[i] = math.NaN()
	}

	// 5. 계산된 SMA 값을 올바른 위치에 복사합니다.
	copy(result[startIdx:], smaValues)

	return result
}

func EmaVWZScores2(zscores []float64, period int) []float64 {
	if len(zscores) < period {
		return nil
	}

	// 초기 SMA로 시작
	sma := talib.Sma(zscores, period)

	fmt.Printf("zscores: %v\n", zscores)
	fmt.Printf("ema: %v\n", sma)

	return sma
}

// VWZScores function as provided by the user
func VWZScores(candles CandleSticks, period int, minStdDev float64) []float64 {
	if len(candles) < period {
		return nil
	}

	vwz := make([]float64, len(candles))

	for i := range candles {
		if i < period-1 {
			vwz[i] = math.NaN()
			continue
		}

		// 기간 슬라이스
		window := candles[i-period+1 : i+1]

		// 가중 평균
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

		// 가중 표준편차
		var variance float64
		for _, c := range window {
			diff := c.Close - mean
			variance += c.Vol * diff * diff
		}
		std := math.Sqrt(variance / weightSum)

		// Z-Score
		if std < minStdDev {
			vwz[i] = math.NaN()
		} else {
			vwz[i] = (candles[i].Close - mean) / std
		}
	}

	return vwz
}

// This is a corrected and efficient version of AdaptiveVWZScoresWithEMA.
// The original function was not suitable for calculating a series where ADX changes at each point.
// This version iterates through the data once, adjusting the EMA smoothing factor 'alpha'
// at each step based on the corresponding ADX value for that point in time.
func calculateAdaptiveVWZScores(
	candles CandleSticks,
	adxSeries []float64,
	adxPeriod int,
	minADX float64,
	maxADX float64,
) []float64 {
	if len(candles) == 0 || len(candles) != len(adxSeries) {
		return nil
	}

	vwz := make([]float64, len(candles))
	minPeriod := adxPeriod - 6 // strong trend → fast mean
	maxPeriod := adxPeriod + 6 // weak trend → slow mean

	var ema, emaSq float64

	for i := range candles {
		if i == 0 {
			ema = candles[i].Close
			emaSq = candles[i].Close * candles[i].Close
			vwz[i] = math.NaN()
			continue
		}

		currentADX := adxSeries[i]
		if math.IsNaN(currentADX) {
			vwz[i] = math.NaN()
			// Keep using previous ema values with a fallback alpha
			ema = (ema * 0.9) + (candles[i].Close * 0.1)
			emaSq = (emaSq * 0.9) + (candles[i].Close * candles[i].Close * 0.1)
			continue
		}

		// ① ADX 값 클램프
		scale := math.Min(math.Max(currentADX, minADX), maxADX)

		// ② 정규화 비율
		ratio := (scale - minADX) / (maxADX - minADX)

		// ③ ADX에 따라 adaptive period 계산
		adaptivePeriod := float64(maxPeriod) - ratio*float64(maxPeriod-minPeriod)

		// ④ smoothing factor
		alpha := 2.0 / (adaptivePeriod + 1.0)

		close := candles[i].Close
		ema = (1-alpha)*ema + alpha*close
		emaSq = (1-alpha)*emaSq + alpha*close*close

		variance := emaSq - ema*ema
		if variance < 0 {
			variance = 0
		}
		std := math.Sqrt(variance)

		if std == 0 {
			vwz[i] = math.NaN()
		} else {
			vwz[i] = (close - ema) / std
		}
	}

	return vwz
}

func BoxFilter(candles CandleSticks, period int, minRangePct float64) []bool {
	// 2. 전체 캔들 데이터 길이만큼 슬라이스를 생성합니다.
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))

	// 3. 데이터를 잘라내지 말고, 전체 캔들을 순회하며 슬라이스를 채웁니다.

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
func ZScores(candles CandleSticks, period int) []float64 {
	if len(candles) < period {
		return nil
	}

	// 종가 추출
	data := make([]float64, 0, len(candles))
	for _, c := range candles {
		data = append(data, c.Close)
	}

	// 이동평균
	mean := talib.Ma(data, period, talib.SMA)

	// 이동 표준편차
	std := talib.StdDev(data, period, 1.0)

	// Z-Score 계산
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
func GetPositionType(longCondition, shortCondition bool) string {
	if longCondition {
		return "LONG"
	}
	if shortCondition {
		return "SHORT"
	}
	return ""
}

// initializeStrategyDataContext는 설정을 기반으로 데이터 로딩 및 모든 기술 지표 계산을 수행하고,
// 완성된 StrategyDataContext를 반환합니다.
func initializeStrategyDataContext(config *Config) (*StrategyDataContext, error) {
	// 1. Read Candles from CSV
	candles, err := readCandlesFromCSV(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read candle data: %w", err)
	}

	if len(candles) == 0 {
		return &StrategyDataContext{Candles: candles}, nil // Return empty context if no candles
	}

	// 2. Prepare data for TALib
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}

	// 3. Calculate all indicator series
	emaShort := talib.Ema(closes, 12)
	emaLong := talib.Ema(closes, 120)
	zScores := ZScores(candles, config.VWZPeriod)
	vwzScores := VWZScores(candles, config.VWZPeriod, config.VWZScore.MinStdDev)
	bbw, _, _, _ := BBW(candles, 20, 2.0)
	bbwzScores := NormalizeBBW(bbw, 50)

	adxSeries := talib.Adx(highs, lows, closes, config.ADXPeriod)
	plusDI := talib.PlusDI(highs, lows, closes, config.ADXPeriod)
	minusDI := talib.MinusDI(highs, lows, closes, config.ADXPeriod)
	dx := talib.Dx(highs, lows, closes, config.ADXPeriod)

	// 4. Create and return the context
	return &StrategyDataContext{
		Candles:    candles,
		EmaShort:   emaShort,
		EmaLong:    emaLong,
		ZScores:    zScores,
		VwzScores:  vwzScores,
		PlusDI:     plusDI,
		MinusDI:    minusDI,
		AdxSeries:  adxSeries,
		BbwzScores: bbwzScores,
		Dx:         dx,
	}, nil
}
