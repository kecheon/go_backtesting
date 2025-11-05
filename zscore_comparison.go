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
func generateHTMLChart(candles CandleSticks, vwzScores []float64, adaptiveVwzScores []float64) {
	var candleData []string
	var vwzData []string
	var adaptiveData []string

	for i, c := range candles {
		ms := c.Time.UnixNano() / int64(time.Millisecond)
		candlePoint := fmt.Sprintf("{x: %d, o: %.4f, h: %.4f, l: %.4f, c: %.4f}", ms, c.Open, c.High, c.Low, c.Close)
		candleData = append(candleData, candlePoint)

		if math.IsNaN(vwzScores[i]) {
			vwzData = append(vwzData, fmt.Sprintf("{x: %d, y: null}", ms))
		} else {
			vwzData = append(vwzData, fmt.Sprintf("{x: %d, y: %.4f}", ms, vwzScores[i]))
		}
		if math.IsNaN(adaptiveVwzScores[i]) {
			adaptiveData = append(adaptiveData, fmt.Sprintf("{x: %d, y: null}", ms))
		} else {
			adaptiveData = append(adaptiveData, fmt.Sprintf("{x: %d, y: %.4f}", ms, adaptiveVwzScores[i]))
		}
	}

	candleDataJS := "[" + strings.Join(candleData, ",") + "]"
	vwzDataJS := "[" + strings.Join(vwzData, ",") + "]"
	adaptiveDataJS := "[" + strings.Join(adaptiveData, ",") + "]"

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Synchronized + Zoomable Candlestick & Z-Score</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-chart-financial"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-zoom@2.0.1"></script>
</head>
<body>
    <canvas id="candleChart" width="1600" height="500"></canvas>
    <canvas id="zscoreChart" width="1600" height="400"></canvas>
    <script>
        const candleData = %s;
        const vwzData = %s;
        const adaptiveVwzData = %s;

        const crosshairPlugin = {
            id: 'crosshair',
            afterDraw: (chart) => {
                if (chart.tooltip?._active?.length) {
                    const ctx = chart.ctx;
                    const x = chart.tooltip._active[0].element.x;
                    const topY = chart.chartArea.top;
                    const bottomY = chart.chartArea.bottom;
                    ctx.save();
                    ctx.beginPath();
                    ctx.moveTo(x, topY);
                    ctx.lineTo(x, bottomY);
                    ctx.lineWidth = 1;
                    ctx.strokeStyle = 'rgba(0, 0, 0, 0.3)';
                    ctx.stroke();
                    ctx.restore();
                }
            }
        };
        Chart.register(crosshairPlugin);

        const ctxCandle = document.getElementById('candleChart').getContext('2d');
        const ctxZScore = document.getElementById('zscoreChart').getContext('2d');

        const commonZoom = {
            zoom: {
                wheel: { enabled: true },
                pinch: { enabled: true },
                drag: { enabled: true },
                mode: 'x'
            },
            pan: {
                enabled: true,
                mode: 'x'
            },
            limits: {
                x: { minRange: 1000 * 60 * 5 } // 최소 5분
            }
        };

        const candleChart = new Chart(ctxCandle, {
            type: 'candlestick',
            data: {
                datasets: [{
                    label: 'SOL/USDT',
                    data: candleData,
                }]
            },
            options: {
                interaction: { intersect: false, mode: 'index' },
                plugins: {
                    legend: { display: true, position: 'top' },
                    zoom: commonZoom
                },
                scales: {
                    x: { type: 'time', time: { unit: 'minute' } },
                    y: { beginAtZero: false }
                }
            }
        });

        const zscoreChart = new Chart(ctxZScore, {
            type: 'line',
            data: {
                datasets: [{
                    label: 'VWZScore',
                    data: vwzData,
                    borderColor: 'rgb(255, 99, 132)',
                    tension: 0.1, pointRadius: 0
                }, {
                    label: 'Adaptive VWZScore',
                    data: adaptiveVwzData,
                    borderColor: 'rgb(54, 162, 235)',
                    tension: 0.1, pointRadius: 0
                }]
            },
            options: {
                interaction: { intersect: false, mode: 'index' },
                plugins: {
                    legend: { display: true, position: 'top' },
                    zoom: commonZoom
                },
                scales: {
                    x: { type: 'time', time: { unit: 'minute' } },
                    y: { beginAtZero: false }
                }
            }
        });

        // 🔄 두 차트 동기화
        function syncCharts(sourceChart, targetChart, event) {
            const points = sourceChart.getElementsAtEventForMode(event, 'index', { intersect: false }, false);
            if (points.length) {
                const index = points[0].index;
                targetChart.setActiveElements([{ datasetIndex: 0, index }]);
                targetChart.tooltip.setActiveElements([{ datasetIndex: 0, index }], {x: 0, y: 0});
                targetChart.update();
            } else {
                targetChart.setActiveElements([]);
                targetChart.tooltip.setActiveElements([], {x: 0, y: 0});
                targetChart.update();
            }
        }

        document.getElementById('candleChart').addEventListener('mousemove', (e) => syncCharts(candleChart, zscoreChart, e));
        document.getElementById('zscoreChart').addEventListener('mousemove', (e) => syncCharts(zscoreChart, candleChart, e));

        document.getElementById('candleChart').addEventListener('mouseleave', () => {
            candleChart.setActiveElements([]); zscoreChart.setActiveElements([]);
            candleChart.tooltip.setActiveElements([], {x: 0, y: 0});
            zscoreChart.tooltip.setActiveElements([], {x: 0, y: 0});
            candleChart.update(); zscoreChart.update();
        });
        document.getElementById('zscoreChart').addEventListener('mouseleave', () => {
            candleChart.setActiveElements([]); zscoreChart.setActiveElements([]);
            candleChart.tooltip.setActiveElements([], {x: 0, y: 0});
            zscoreChart.tooltip.setActiveElements([], {x: 0, y: 0});
            candleChart.update(); zscoreChart.update();
        });

        // 더블클릭으로 줌 리셋
        document.getElementById('candleChart').addEventListener('dblclick', () => {
            candleChart.resetZoom();
            zscoreChart.resetZoom();
        });
        document.getElementById('zscoreChart').addEventListener('dblclick', () => {
            candleChart.resetZoom();
            zscoreChart.resetZoom();
        });
    </script>
</body>
</html>
`, candleDataJS, vwzDataJS, adaptiveDataJS)

	os.WriteFile("chart.html", []byte(htmlContent), 0644)
	fmt.Println("Generated chart.html with zoom & sync (scroll or drag to zoom)")
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
	reader.TrimLeadingSpace = true
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

		open, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Printf("Error parsing open price: %v", err)
			continue
		}
		high, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Error parsing high price: %v", err)
			continue
		}
		low, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Error parsing low price: %v", err)
			continue
		}
		close, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			log.Printf("Error parsing close price: %v", err)
			continue
		}
		vol, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Printf("Error parsing volume: %v", err)
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

	// --- Print Results ---
	fmt.Printf("\n--- Z-Score Comparison ---\n")
	fmt.Println("Comparing VWZScores and Adaptive VWZScores where either Z-Score >= 1.5")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("%-25s %-20s %-20s\n", "Timestamp", "VWZScore", "Adaptive VWZScore")
	fmt.Println("-----------------------------------------------------------------")

	for i := range candles {
		vwz := vwzScores[i]
		adaptiveVwz := adaptiveVwzScores[i]

		if (vwz >= 1.5 && !math.IsNaN(vwz)) && (adaptiveVwz >= 1.5 && !math.IsNaN(adaptiveVwz)) {
			vwzStr := "NaN"
			if !math.IsNaN(vwz) {
				vwzStr = fmt.Sprintf("%.4f", vwz)
			}

			adaptiveVwzStr := "NaN"
			if !math.IsNaN(adaptiveVwz) {
				adaptiveVwzStr = fmt.Sprintf("%.4f", adaptiveVwz)
			}

			fmt.Printf("%-25s %-20s %-20s\n",
				candles[i].Time.Format("2006-01-02 15:04:05"),
				vwzStr,
				adaptiveVwzStr,
			)
		}
	}
	fmt.Println("-----------------------------------------------------------------")

	// --- Generate Chart ---
	generateHTMLChart(candles, vwzScores, adaptiveVwzScores)
}
