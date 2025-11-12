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

const ADXTHRESHOLD = 25.0

// var filePath = "BTC.csv"
var filePath = "SOLUSDT_5m_raw_data.csv"
var cutoffDate = time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC)
var vwzPeriod = 48
var adxPeriod = 14

func main() {
	// --- Configuration ---
	// filePath := "SOLUSDT_5m_raw_data.csv"
	// filePath := "ETHUSDT_5m_raw_data.csv"
	// Adaptive EMA parameters
	// minADX := 20.0
	// maxADX := 50.0
	// Date filter (Assuming the year is 2024, as it's the most recent October)
	// The user can change the year if needed.

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
	vols := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
		vols[i] = c.Vol
	}

	// --- Calculations ---
	// 1. VWZScores
	// zScores := ZScores(candles, vwzPeriod)
	// vwzScores := VWZScores(candles, vwzPeriod)
	zScores := Ema(ZScores(candles, vwzPeriod), 26)
	vwzScores := Ema(VWZScores(candles, vwzPeriod), 26)
	// zScores := ZScores(candles, vwzPeriod)
	// vwzScores := VWZScores(candles, vwzPeriod)
	// emaVwz := EmaVWZScores(vwzScores, 9)
	// longEmaVwz := EmaVWZScores(vwzScores, 25)

	// 2. Adaptive VWZScores
	adxSeries := talib.Adx(highs, lows, closes, adxPeriod)
	plusDI := talib.PlusDI(highs, lows, closes, adxPeriod)
	minusDI := talib.MinusDI(highs, lows, closes, adxPeriod)
	dx := talib.Dx(highs, lows, closes, adxPeriod)
	// adaptiveVwzScores := calculateAdaptiveVWZScores(candles, adxSeries, adxPeriod, minADX, maxADX)

	// --- Print Results ---
	fmt.Printf("\n--- Z-Score Comparison ---\n")
	fmt.Println("Comparing ZScores and VWZScores where either Z-Score >= 1.5")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("%-25s %-15s %-15s %-15s %-15s %-15s %-15s %-15s\n", "Timestamp", "ZScore", "VWZScore", "ADX", "Volume", "PlusDI", "MinusDI", "DX")
	fmt.Println("-----------------------------------------------------------------")

	count := 0
	var entrySignals []EntrySignal
	for i := range candles {
		if i < vwzPeriod-1 || i < adxPeriod-1 {
			continue
		}

		zscores := zScores[i]
		// pzscores := zScores[i-2]
		vwzScore := vwzScores[i]
		pvwzScore := vwzScores[i-10]
		adx := adxSeries[i]
		vol := vols[i]
		plusDI := plusDI[i]
		minusDI := minusDI[i]
		dx := dx[i]

		// condition := (vwz > 0 && pvwz < 0 || vwz < 0 && pvwz > 0) && !math.IsNaN(pvwz)
		// isRanging := BoxFilter(candles)[i]
		// condition := adx > ADXTHRESHOLD && ((math.Abs(zscores) >= 0) && (math.Abs(vwzScore) >= 1.5))
		// crossCondition := vwzScore > 0.3 && pvwzScore < -0.3 || vwzScore < -0.3 && pvwzScore > 0.3
		longCondition := pvwzScore < -0.5 && vwzScore > 0.5
		// longCondition = longCondition && pzscores < -1.5 && zscores > -1.5
		shortCondition := pvwzScore > 0.5 && vwzScore < -0.5
		// shortCondition = shortCondition && pzscores > 1.5 && zscores < 1.5
		condition := dx > 25 && adx > ADXTHRESHOLD && (longCondition || shortCondition)

		if condition {
			var direction string
			if longCondition {
				direction = "long"
			} else {
				direction = "short"
			}
			entrySignals = append(entrySignals, EntrySignal{
				Time:      candles[i].Time,
				Price:     candles[i].Close,
				Direction: direction,
			})

			zStr := "NaN"
			if !math.IsNaN(vwzScore) {
				zStr = fmt.Sprintf("%.4f", zscores)
			}

			vwzStr := "NaN"
			if !math.IsNaN(zscores) {
				vwzStr = fmt.Sprintf("%.4f", vwzScore)
			}

			fmt.Printf("[%d] %s %-25s %-15s %-15s %-15s %-15s %-15s %-15s %-15s\n",
				count,
				GetPositionType(longCondition, shortCondition),
				candles[i].Time.Format("01-02 15:04"),
				zStr,
				vwzStr,
				fmt.Sprintf("%.2f", adx),
				fmt.Sprintf("%.2f", vol),
				fmt.Sprintf("%.2f", plusDI),
				fmt.Sprintf("%.2f", minusDI),
				fmt.Sprintf("%.2f", dx),
			)
			count = count + 1
		}
	}
	fmt.Println("-----------------------------------------------------------------")

	// --- Generate Chart ---
	generateHTMLChart(candles, zScores, vwzScores, entrySignals)
}
