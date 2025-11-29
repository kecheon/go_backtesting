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

	// --- 3. Get Entry Conditions ---
	longCondition, err := strategy.GetEntryCondition(cfg.LongCondition, "long")
	if err != nil {
		log.Fatalf("Failed to get long entry condition: %v", err)
	}

	shortCondition, err := strategy.GetEntryCondition(cfg.ShortCondition, "short")
	if err != nil {
		log.Fatalf("Failed to get short entry condition: %v", err)
	}

	// --- 4. Run Selected Mode ---
	if cfg.RunMode == "signals" {
		// --- Generate and Print All Signals ---
		signals := strategy.GenerateAllSignals(strategyData, cfg, longCondition, shortCondition)
		reporting.PrintAllSignals(signals)
		reporting.GenerateHTMLChart(strategyData.Candles, strategyData.BoxFilter, strategyData.VwzScores, signals)
	} else {
		// --- Run Backtest and Print Results ---
		result := strategy.RunBacktest(strategyData, cfg, longCondition, shortCondition)
		reporting.PrintDetailedTradeRecords(result)
		reporting.PrintTradeAnalysis(result, strategyData)
		reporting.PrintBacktestSummary(result)

		var entrySignals []strategy.EntrySignal
		for _, trade := range result.Trades {
			entrySignals = append(entrySignals, strategy.EntrySignal{
				Time:      trade.EntryTime,
				Price:     trade.EntryPrice,
				Direction: trade.Direction,
			})
		}
		reporting.GenerateHTMLChart(strategyData.Candles, strategyData.BoxFilter, strategyData.VwzScores, entrySignals)
	}
}
