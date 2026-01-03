package strategy

import (
	"go-backtesting/config"
	"time"
)

// EntrySignal represents an entry signal.
type EntrySignal struct {
	Time      time.Time
	Price     float64
	Direction string // "long" or "short"
}

// Trade represents a single trade.
type Trade struct {
	EntryTime       time.Time
	EntryPrice      float64
	ExitTime        time.Time
	ExitPrice       float64
	Direction       string // "long" or "short"
	Pnl             float64
	PnlPercentage   float64
	EntryIndicators TechnicalIndicators
}

// BacktestResult contains the results of a backtest.
type BacktestResult struct {
	Trades      []Trade
	TotalPnl    float64
	WinCount    int
	LossCount   int
	TotalTrades int
	WinRate     float64
}

// RunBacktest runs a backtest and returns the results.
func RunBacktest(strategyData *StrategyDataContext, config *config.Config, longCondition EntryCondition, shortCondition EntryCondition) BacktestResult {
	var activeTrade *Trade
	var completedTrades []Trade

	takeProfitPct := config.TPRate // 1% take profit
	stopLossPct := config.SLRate   // 1% stop loss

	for i := range strategyData.Candles {
		currentCandle := strategyData.Candles[i]

		// --- 1. Exit Logic: Check if there is an active trade ---
		if activeTrade != nil {
			indicators := strategyData.createTechnicalIndicators(i, config)
			direction, entry, stop := DetermineEntrySignal(indicators, config, longCondition, shortCondition)
			isPriceThresholdBreached := false
			var exitPrice float64

			if activeTrade.Direction == "long" {
				takeProfitPrice := activeTrade.EntryPrice * (1 + takeProfitPct)
				stopLossPrice := activeTrade.EntryPrice * (1 - stopLossPct)

				if currentCandle.Low <= stopLossPrice {
					isPriceThresholdBreached = true
					exitPrice = stopLossPrice
				} else if currentCandle.High >= takeProfitPrice {
					isPriceThresholdBreached = true
					exitPrice = takeProfitPrice
				} else if (entry && direction == "short") || (stop && direction == "long") {
					isPriceThresholdBreached = true
					exitPrice = currentCandle.Close
				}

			} else { // short
				takeProfitPrice := activeTrade.EntryPrice * (1 - takeProfitPct)
				stopLossPrice := activeTrade.EntryPrice * (1 + stopLossPct)

				if currentCandle.High >= stopLossPrice {
					isPriceThresholdBreached = true
					exitPrice = stopLossPrice
				} else if currentCandle.Low <= takeProfitPrice {
					isPriceThresholdBreached = true
					exitPrice = takeProfitPrice
				} else if (entry && direction == "long") || (stop && direction == "short") {
					isPriceThresholdBreached = true
					exitPrice = currentCandle.Close
				}
			}

			finalExitTrigger := false
			if isPriceThresholdBreached {
				shouldHold := false
				if activeTrade.Direction == "long" &&
					i < len(strategyData.PlusDI) &&
					i < len(strategyData.MinusDI) &&
					strategyData.PlusDI[i] > strategyData.MinusDI[i] &&
					strategyData.AdxSeries[i-1] < strategyData.AdxSeries[i] {
					shouldHold = false
				} else if activeTrade.Direction == "short" &&
					i < len(strategyData.PlusDI) &&
					i < len(strategyData.MinusDI) &&
					strategyData.AdxSeries[i-1] < strategyData.AdxSeries[i] &&
					strategyData.MinusDI[i] > strategyData.PlusDI[i] {
					shouldHold = false
				}

				if !shouldHold {
					finalExitTrigger = true
				}
			}

			if finalExitTrigger {
				// Use the specific exit price calculated above
				activeTrade.ExitTime = currentCandle.Time
				activeTrade.ExitPrice = exitPrice
				if activeTrade.Direction == "long" {
					activeTrade.Pnl = activeTrade.ExitPrice - activeTrade.EntryPrice
				} else {
					activeTrade.Pnl = activeTrade.EntryPrice - activeTrade.ExitPrice
				}
				activeTrade.PnlPercentage = (activeTrade.Pnl / activeTrade.EntryPrice) * 100
				completedTrades = append(completedTrades, *activeTrade)
				activeTrade = nil // Close the position
			}
		}

		// --- 2. Entry Logic: Only enter if there is no active trade ---
		if activeTrade == nil {
			if i < 14 { // Hardcoded ATR period in DetectBBWState
				continue
			}
			if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
				continue
			}
			indicators := strategyData.createTechnicalIndicators(i, config)
			direction, entry, _ := DetermineEntrySignal(indicators, config, longCondition, shortCondition)

			if entry {
				activeTrade = &Trade{
					EntryTime:       currentCandle.Time,
					EntryPrice:      currentCandle.Close,
					Direction:       direction,
					EntryIndicators: indicators,
				}
			}
		}
	}

	// --- 3. Final Result Calculation ---
	var totalPnl float64
	winCount := 0
	lossCount := 0
	for _, t := range completedTrades {
		totalPnl += t.Pnl
		if t.Pnl > 0 {
			winCount++
		} else {
			lossCount++
		}
	}

	totalTrades := len(completedTrades)
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winCount) / float64(totalTrades) * 100
	}

	return BacktestResult{
		Trades:      completedTrades,
		TotalPnl:    totalPnl,
		WinCount:    winCount,
		LossCount:   lossCount,
		TotalTrades: totalTrades,
		WinRate:     winRate,
	}
}

// GenerateAllSignals iterates through the strategy data and returns all entry signals.
func GenerateAllSignals(strategyData *StrategyDataContext, config *config.Config, longCondition EntryCondition, shortCondition EntryCondition) []EntrySignal {
	var signals []EntrySignal

	for i := range strategyData.Candles {
		if i < 14 { // Hardcoded ATR period in DetectBBWState
			continue
		}
		if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
			continue
		}

		indicators := strategyData.createTechnicalIndicators(i, config)
		direction, entry, _ := DetermineEntrySignal(indicators, config, longCondition, shortCondition)

		if entry {
			signal := EntrySignal{
				Time:      strategyData.Candles[i].Time,
				Price:     strategyData.Candles[i].Close,
				Direction: direction,
			}
			signals = append(signals, signal)
		}
	}

	return signals
}
