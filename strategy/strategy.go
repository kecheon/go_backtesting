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
			isPriceThresholdBreached := false
			if activeTrade.Direction == "long" {
				takeProfitPrice := activeTrade.EntryPrice * (1 + takeProfitPct)
				stopLossPrice := activeTrade.EntryPrice * (1 - stopLossPct)
				if currentCandle.High >= takeProfitPrice || currentCandle.Low <= stopLossPrice {
					isPriceThresholdBreached = true
				}
			} else { // short
				takeProfitPrice := activeTrade.EntryPrice * (1 - takeProfitPct)
				stopLossPrice := activeTrade.EntryPrice * (1 + stopLossPct)
				if currentCandle.Low <= takeProfitPrice || currentCandle.High >= stopLossPrice {
					isPriceThresholdBreached = true
				}
			}

			finalExitTrigger := false
			if isPriceThresholdBreached {
				shouldHold := false
				if activeTrade.Direction == "long" &&
					i < len(strategyData.PlusDI) &&
					i < len(strategyData.MinusDI) &&
					strategyData.PlusDI[i] > strategyData.MinusDI[i] {
					shouldHold = false
				} else if activeTrade.Direction == "short" &&
					i < len(strategyData.PlusDI) &&
					i < len(strategyData.MinusDI) &&
					strategyData.MinusDI[i] > strategyData.PlusDI[i] {
					shouldHold = false
				}

				if !shouldHold {
					finalExitTrigger = true
				}
			}

			if finalExitTrigger {
				exitPrice := currentCandle.Close
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
			if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
				continue
			}
			indicators := strategyData.createTechnicalIndicators(i, config)
			direction, hasSignal := DetermineEntrySignal(indicators, config.ADXThreshold, longCondition, shortCondition)

			if hasSignal {
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

// EntryCondition defines the signature for a function that checks for a trading signal.
type EntryCondition func(indicators TechnicalIndicators) bool

// DetermineEntrySignal determines the entry signal based on the indicators.
func DetermineEntrySignal(indicators TechnicalIndicators, adxThreshold float64, longCondition EntryCondition, shortCondition EntryCondition) (string, bool) {
	if indicators.ADX > adxThreshold && indicators.ADX < 50 {
		if longCondition(indicators) {
			return "long", true
		}
		if shortCondition(indicators) {
			return "short", true
		}
	}
	return "", false
}

// DefaultLongCondition provides the default logic for a long entry signal.
func DefaultLongCondition(indicators TechnicalIndicators) bool {
	return indicators.EmaShort > indicators.EmaLong && indicators.ZScore < 0.0
}

// DefaultShortCondition provides the default logic for a short entry signal.
func DefaultShortCondition(indicators TechnicalIndicators) bool {
	return indicators.EmaShort < indicators.EmaLong && indicators.ZScore > 0.0
}
