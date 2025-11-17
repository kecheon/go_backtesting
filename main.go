package main

import (
	"log"

	"go-backtesting/config"
	"go-backtesting/reporting"
	"go-backtesting/strategy"
)


func main() {
	// --- 1. Load Configuration ---
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- 2. Initialize All Strategy Data ---
	strategyData, err := strategy.InitializeStrategyDataContext(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize strategy data: %v", err)
	}

	if len(strategyData.Candles) == 0 {
		log.Println("No data available for the specified date range.")
		return
	}

	// --- 3. Run Backtest ---
	result := strategy.RunBacktest(strategyData, cfg, strategy.DefaultLongCondition, strategy.DefaultShortCondition)

	// --- 4. Print Reports and Generate Chart ---
	reporting.PrintDetailedTradeRecords(result)
	reporting.PrintTradeAnalysis(result, strategyData)
	reporting.PrintBacktestSummary(result)

	// Create EntrySignal slice from actual trades for chart generation
	var entrySignals []strategy.EntrySignal
	for _, trade := range result.Trades {
		entrySignals = append(entrySignals, strategy.EntrySignal{
			Time:      trade.EntryTime,
			Price:     trade.EntryPrice,
			Direction: trade.Direction,
		})
	}

	// Generate the chart with signals that were actually traded
	reporting.GenerateHTMLChart(strategyData.Candles, strategyData.ZScores, strategyData.VwzScores, entrySignals)
}
