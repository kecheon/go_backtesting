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

// printTradeAnalysis는 백테스트에서 실제 거래된 내역을 기반으로 신호 분석표를 출력합니다.
func printTradeAnalysis(result BacktestResult, strategyData *StrategyDataContext) {
	fmt.Printf("\n--- Trade Entry Analysis ---\n")
	fmt.Println("Signals that resulted in a trade:")
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-5s %-5s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "Idx", "Type", "Timestamp", "ZScore", "VWZScore", "BBW", "ADX", "Volume", "PlusDI", "MinusDI", "DX", "PnL", "PnL(%)")
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------")

	for i, trade := range result.Trades {
		indicators := trade.EntryIndicators
		zStr := "NaN"
		if !math.IsNaN(indicators.VWZScore) {
			zStr = fmt.Sprintf("%.4f", indicators.ZScore)
		}

		vwzStr := "NaN"
		if !math.IsNaN(indicators.ZScore) {
			vwzStr = fmt.Sprintf("%.4f", indicators.VWZScore)
		}

		entryIndex := -1
		for j, c := range strategyData.Candles {
			if c.Time == trade.EntryTime {
				entryIndex = j
				break
			}
		}

		bbwStr := "NaN"
		if entryIndex != -1 && entryIndex < len(strategyData.BbwzScores) && !math.IsNaN(strategyData.BbwzScores[entryIndex]) {
			bbwStr = fmt.Sprintf("%.4f", strategyData.BbwzScores[entryIndex])
		}
		dxStr := "NaN"
		if entryIndex != -1 && entryIndex < len(strategyData.DX) && !math.IsNaN(strategyData.DX[entryIndex]) {
			dxStr = fmt.Sprintf("%.2f", strategyData.DX[entryIndex])
		}

		volStr := "NaN"
		if entryIndex != -1 {
			volStr = fmt.Sprintf("%.2f", strategyData.Candles[entryIndex].Vol)
		}

		pnlPctStr := fmt.Sprintf("%.2f%%", trade.PnlPercentage)

		fmt.Printf("%-5d %-5s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10.2f %-10s\n",
			i,
			trade.Direction,
			trade.EntryTime.Format("01-02 15:04"),
			zStr,
			vwzStr,
			bbwStr,
			fmt.Sprintf("%.2f", indicators.ADX),
			volStr,
			fmt.Sprintf("%.2f", indicators.PlusDI),
			fmt.Sprintf("%.2f", indicators.MinusDI),
			dxStr,
			trade.Pnl,
			pnlPctStr,
		)
	}
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------")
}

func main() {
	// --- 1. Load Configuration ---
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- 2. Initialize All Strategy Data ---
	strategyData, err := initializeStrategyDataContext(config)
	if err != nil {
		log.Fatalf("Failed to initialize strategy data: %v", err)
	}

	if len(strategyData.Candles) == 0 {
		log.Println("No data available for the specified date range.")
		return
	}

	// --- 3. Run Backtest ---
	result := runBacktest(strategyData, config)

	// --- 4. Print Reports ---
	printDetailedTradeRecords(result)
	printTradeAnalysis(result, strategyData)
	printBacktestSummary(result)

	// The chart generation is commented out as it might need adaptation
	// generateHTMLChart(strategyData.Candles, strategyData.ZScores, strategyData.VwzScores, entrySignals)
}
