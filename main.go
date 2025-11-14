package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
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

	// --- Initialize All Strategy Data ---
	strategyData, err := initializeStrategyDataContext(config)
	if err != nil {
		log.Fatalf("Failed to initialize strategy data: %v", err)
	}

	if len(strategyData.Candles) == 0 {
		log.Println("No data available for the specified date range.")
		return
	}

	// --- Print Results ---
	fmt.Printf("\n--- Z-Score Comparison ---\n")
	fmt.Println("Comparing ZScores and VWZScores")
	fmt.Println("-----------------------------------------------------------------")
	// The first two columns ([count] and Position Type) are printed without a header.
	fmt.Printf("%-5s %-5s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "Idx", "Type", "Timestamp", "ZScore", "VWZScore", "BBW", "ADX", "Volume", "PlusDI", "MinusDI", "DX")
	fmt.Println("-----------------------------------------------------------------")

	count := 0
	var entrySignals []EntrySignal
	for i := range strategyData.Candles {
		if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
			continue
		}

		// --- 현재 인덱스(i)에 대한 기술 지표 구조체 생성 ---
		indicators := strategyData.createTechnicalIndicators(i)

		// --- 진입 신호 결정 ---
		direction, hasSignal := determineEntrySignal(indicators, config.ADXThreshold)

		if hasSignal {
			entrySignals = append(entrySignals, EntrySignal{
				Time:      strategyData.Candles[i].Time,
				Price:     strategyData.Candles[i].Close,
				Direction: direction,
			})

			zStr := "NaN"
			if !math.IsNaN(indicators.VWZScore) {
				zStr = fmt.Sprintf("%.4f", indicators.ZScore)
			}

			vwzStr := "NaN"
			if !math.IsNaN(indicators.ZScore) {
				vwzStr = fmt.Sprintf("%.4f", indicators.VWZScore)
			}
			bbwStr := "NaN"
			if i < len(strategyData.BbwzScores) && !math.IsNaN(strategyData.Bbw[i]) {
				bbwStr = fmt.Sprintf("%.4f", strategyData.Bbw[i])
			}
			dxStr := "NaN"
			if i < len(strategyData.Dx) && !math.IsNaN(strategyData.Dx[i]) {
				dxStr = fmt.Sprintf("%.2f", strategyData.Dx[i])
			}

			fmt.Printf("%-5d %-5s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n",
				count,
				GetPositionType(direction == "long", direction == "short"),
				strategyData.Candles[i].Time.Format("01-02 15:04"),
				zStr,
				vwzStr,
				bbwStr,
				fmt.Sprintf("%.2f", indicators.ADX),
				fmt.Sprintf("%.2f", strategyData.Candles[i].Vol),
				fmt.Sprintf("%.2f", indicators.PlusDI),
				fmt.Sprintf("%.2f", indicators.MinusDI),
				dxStr,
			)
			count = count + 1
		}
	}
	fmt.Println("-----------------------------------------------------------------")

	// --- Generate Chart ---
	generateHTMLChart(strategyData.Candles, strategyData.ZScores, strategyData.VwzScores, entrySignals)
}
