package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/markcheno/go-talib"
)

// --- Configuration Structs ---
type BoxFilterConfig struct {
	Period      int     `json:"period"`
	MinRangePct float64 `json:"minRangePct"`
}

type VWZScoreConfig struct {
	MinStdDev float64 `json:"minStdDev"`
}

type Config struct {
	FilePath        string          `json:"filePath"`
	VWZPeriod       int             `json:"vwzPeriod"`
	ZScoreThreshold float64         `json:"zscoreThreshold"`
	EmaPeriod       int             `json:"emaPeriod"`
	BoxFilter       BoxFilterConfig `json:"boxFilter"`
	VWZScore        VWZScoreConfig  `json:"vwzScore"`
	ADXPeriod       int             `json:"adxPeriod"`
	ADXThreshold    float64         `json:"adxThreshold"`
}

// --- Global Config Variable ---
var config *Config

// loadConfig reads and parses the configuration file.
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	cfg := &Config{}
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}
	return cfg, nil
}

func main() {
	// --- Load Configuration ---
	var err error
	config, err = loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- Read and Parse CSV ---
	file, err := os.Open(config.FilePath)
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
	emaShort := talib.Ema(closes, 12)
	emaLong := talib.Ema(closes, 120)
	// zScores := Ema(ZScores(candles, config.VWZPeriod), config.EmaPeriod)
	// vwzScores := Ema(VWZScores(candles, config.VWZPeriod, config.VWZScore.MinStdDev), config.EmaPeriod)
	zScores := ZScores(candles, config.VWZPeriod)
	vwzScores := VWZScores(candles, config.VWZPeriod, config.VWZScore.MinStdDev)
	bbw, _, _, _ := BBW(candles, 20, 2.0)
	bbwzScores := NormalizeBBW(bbw, 50)

	adxSeries := talib.Adx(highs, lows, closes, config.ADXPeriod)
	plusDI := talib.PlusDI(highs, lows, closes, config.ADXPeriod)
	minusDI := talib.MinusDI(highs, lows, closes, config.ADXPeriod)
	dx := talib.Dx(highs, lows, closes, config.ADXPeriod)

	// --- Print Results ---
	fmt.Printf("\n--- Z-Score Comparison ---\n")
	fmt.Println("Comparing ZScores and VWZScores")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("%-30s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "Timestamp", "ZScore", "VWZScore", "BBW", "ADX", "Volume", "PlusDI", "MinusDI", "DX")
	fmt.Println("-----------------------------------------------------------------")

	count := 0
	var entrySignals []EntrySignal
	for i := range candles {
		if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
			continue
		}

		ema_short := emaShort[i]
		ema_long := emaLong[i]

		zscores := zScores[i]
		vwzScore := vwzScores[i]
		bbwzScore := bbwzScores[i]
		// pvwzScore := vwzScores[i-10]
		adx := adxSeries[i]
		vol := vols[i]
		plusDI := plusDI[i]
		minusDI := minusDI[i]
		dx := dx[i]

		bbState := DetectBBWState(candles[:i], 20, 2.0, 0)

		// if bbState.Status == ExpandingBearish && bbwScore < -1.5 {
		// 	fmt.Println("++++++++++++++++++++++++")
		// 	fmt.Printf("bbwScore: %.2f minusDI %.2f plusDI %.2f vwzScore %.2f\n", bbwScore, minusDI, plusDI, vwzScore)
		// }

		longCondition := bbState.Status == ExpandingBullish && plusDI > minusDI && vwzScore < 1.0 && zscores < 1.0 && ema_short > ema_long
		shortCondition := bbState.Status == ExpandingBearish && minusDI > plusDI && vwzScore > -1.0 && zscores > -1.0 && ema_short < ema_long
		condition := adx > config.ADXThreshold && (longCondition || shortCondition)

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
			bbwStr := "NaN"
			if !math.IsNaN(bbwzScore) {
				bbwStr = fmt.Sprintf("%.4f", bbwzScore)
			}

			fmt.Printf("[%d] %s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %-15s %-10s\n",
				count,
				GetPositionType(longCondition, shortCondition),
				candles[i].Time.Format("01-02 15:04"),
				zStr,
				vwzStr,
				bbwStr,
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
