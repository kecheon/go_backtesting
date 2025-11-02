package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
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

// VWZScores function as provided by the user
func VWZScores(candles CandleSticks, period int) []float64 {
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
		if std == 0 {
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
	basePeriod := adxPeriod + 4
	minPeriod := basePeriod / 2 // strong trend → fast mean
	maxPeriod := basePeriod * 2 // weak trend → slow mean

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
			ema = (ema*0.9) + (candles[i].Close * 0.1)
			emaSq = (emaSq*0.9) + (candles[i].Close*candles[i].Close*0.1)
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

func generateHTMLChart(candles CandleSticks, vwzScores []float64, adaptiveVwzScores []float64) {
	// Prepare data for the template
	var labels []string
	var vwzData []string
	var adaptiveData []string

	for i, c := range candles {
		labels = append(labels, c.Time.Format("2006-01-02 15:04:05"))

		if math.IsNaN(vwzScores[i]) {
			vwzData = append(vwzData, "null")
		} else {
			vwzData = append(vwzData, fmt.Sprintf("%.4f", vwzScores[i]))
		}

		if math.IsNaN(adaptiveVwzScores[i]) {
			adaptiveData = append(adaptiveData, "null")
		} else {
			adaptiveData = append(adaptiveData, fmt.Sprintf("%.4f", adaptiveVwzScores[i]))
		}
	}

	// Using strings.Join for creating JS arrays.
	labelsJS := "['" + strings.Join(labels, "','") + "']"
	vwzDataJS := "[" + strings.Join(vwzData, ",") + "]"
	adaptiveDataJS := "[" + strings.Join(adaptiveData, ",") + "]"

	// HTML content
	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Z-Score Comparison Chart</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <canvas id="zscoreChart" width="1600" height="900"></canvas>
    <script>
        const ctx = document.getElementById('zscoreChart').getContext('2d');
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: %s,
                datasets: [{
                    label: 'VWZScore',
                    data: %s,
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.5)',
                    tension: 0.1
                }, {
                    label: 'Adaptive VWZScore',
                    data: %s,
                    borderColor: 'rgb(54, 162, 235)',
                    backgroundColor: 'rgba(54, 162, 235, 0.5)',
                    tension: 0.1
                }]
            },
            options: {
                scales: {
                    y: {
                        beginAtZero: false
                    }
                }
            }
        });
    </script>
</body>
</html>
`, labelsJS, vwzDataJS, adaptiveDataJS)

	// Write the content to a file
	err := os.WriteFile("chart.html", []byte(htmlContent), 0644)
	if err != nil {
		log.Fatalf("Error writing chart.html file: %v", err)
	}

	fmt.Println("Generated chart.html")
}

func main() {
	// --- Configuration ---
	filePath := "SOLUSDT_5m_raw_data.csv"
	vwzPeriod := 20
	adxPeriod := 14
	// Adaptive EMA parameters
	minADX := 20.0
	maxADX := 50.0
	// Date filter (Assuming the year is 2024, as it's the most recent October)
	// The user can change the year if needed.
	cutoffDate := time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC)

	// --- Read and Parse CSV ---
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read() // Skip header
	if err != nil {
		log.Fatalf("Error reading header: %v", err)
	}

	var candles CandleSticks
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading record: %v", err)
			continue
		}

		// Assuming CSV format: YYYY-MM-DD HH:MM:SS, open, high, low, close, volume
		t, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			continue
		}

		if t.Before(cutoffDate) {
			continue
		}

		open, _ := strconv.ParseFloat(record[1], 64)
		high, _ := strconv.ParseFloat(record[2], 64)
		low, _ := strconv.ParseFloat(record[3], 64)
		close, _ := strconv.ParseFloat(record[4], 64)
		vol, _ := strconv.ParseFloat(record[5], 64)

		candles = append(candles, Candle{
			Time:  t,
			Open:  open,
			High:  high,
			Low:   low,
			Close: close,
			Vol:   vol,
		})
	}

	if len(candles) == 0 {
		log.Println("No data available for the specified date range.")
		return
	}

	// --- Prepare data for TALib ---
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}

	// --- Calculations ---
	// 1. VWZScores
	vwzScores := VWZScores(candles, vwzPeriod)

	// 2. Adaptive VWZScores
	adxSeries := talib.Adx(highs, lows, closes, adxPeriod)
	adaptiveVwzScores := calculateAdaptiveVWZScores(candles, adxSeries, adxPeriod, minADX, maxADX)

	// --- Generate Chart ---
	generateHTMLChart(candles, vwzScores, adaptiveVwzScores)
}