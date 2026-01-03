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
// Hedge Mode Implementation: Supports simultaneous Long and Short positions.
func RunBacktest(strategyData *StrategyDataContext, config *config.Config, longCondition EntryCondition, shortCondition EntryCondition) BacktestResult {
	var activeLongTrade *Trade
	var activeShortTrade *Trade
	var completedTrades []Trade

	takeProfitPct := config.TPRate
	stopLossPct := config.SLRate

	for i := range strategyData.Candles {
		currentCandle := strategyData.Candles[i]

		// 0. Hedge Exit Logic: If both positions are open, check Combined PnL > 0
		if activeLongTrade != nil && activeShortTrade != nil {
			longPnl := (currentCandle.Close - activeLongTrade.EntryPrice)
			shortPnl := (activeShortTrade.EntryPrice - currentCandle.Close)
			combinedPnl := longPnl + shortPnl

			if combinedPnl > 0 {
				// Close Both
				// Long
				activeLongTrade.ExitTime = currentCandle.Time
				activeLongTrade.ExitPrice = currentCandle.Close
				activeLongTrade.Pnl = longPnl
				activeLongTrade.PnlPercentage = (activeLongTrade.Pnl / activeLongTrade.EntryPrice) * 100
				completedTrades = append(completedTrades, *activeLongTrade)
				activeLongTrade = nil

				// Short
				activeShortTrade.ExitTime = currentCandle.Time
				activeShortTrade.ExitPrice = currentCandle.Close
				activeShortTrade.Pnl = shortPnl
				activeShortTrade.PnlPercentage = (activeShortTrade.Pnl / activeShortTrade.EntryPrice) * 100
				completedTrades = append(completedTrades, *activeShortTrade)
				activeShortTrade = nil

				continue // Move to next candle
			}
		}

		indicators := strategyData.createTechnicalIndicators(i, config)
		_, longEntry, longStop := DetermineEntrySignal(indicators, config, longCondition, shortCondition) // "long", entry, stop is handled differently?
		// DetermineEntrySignal returns (direction, entry, stop).
		// We need independent checks for Long and Short logic.
		// Actually DetermineEntrySignal checks both conditions and returns the FIRST one that triggers.
		// In Hedge Mode, we want to evaluate BOTH.
		// So we should call longCondition and shortCondition directly?
		// Yes, DetermineEntrySignal is for Single Position logic mostly.
		// Or we can assume DetermineEntrySignal prioritizes Long?
		// Let's call conditions directly for full control.

		isLongEntry, isLongStop := longCondition(indicators, config)
		isShortEntry, isShortStop := shortCondition(indicators, config)

		// --- 1. Long Trade Management ---
		if activeLongTrade != nil {
			exitPrice := 0.0
			closed := false

			takeProfitPrice := activeLongTrade.EntryPrice * (1 + takeProfitPct)
			stopLossPrice := activeLongTrade.EntryPrice * (1 - stopLossPct)

			// Check TP/SL
			if currentCandle.High >= takeProfitPrice {
				exitPrice = takeProfitPrice
				closed = true
			} else if currentCandle.Low <= stopLossPrice {
				exitPrice = stopLossPrice
				closed = true
			} else if isLongStop { // Signal Exit
				exitPrice = currentCandle.Close
				closed = true
			}

			if closed {
				activeLongTrade.ExitTime = currentCandle.Time
				activeLongTrade.ExitPrice = exitPrice
				activeLongTrade.Pnl = activeLongTrade.ExitPrice - activeLongTrade.EntryPrice
				activeLongTrade.PnlPercentage = (activeLongTrade.Pnl / activeLongTrade.EntryPrice) * 100
				completedTrades = append(completedTrades, *activeLongTrade)
				activeLongTrade = nil
			}
		} else {
			// Long Entry
			// Wait for warm up
			if i >= 14 && i >= config.VWZPeriod-1 && i >= config.ADXPeriod-1 {
				if isLongEntry {
					activeLongTrade = &Trade{
						EntryTime:       currentCandle.Time,
						EntryPrice:      currentCandle.Close,
						Direction:       "long",
						EntryIndicators: indicators,
					}
				}
			}
		}

		// --- 2. Short Trade Management ---
		if activeShortTrade != nil {
			exitPrice := 0.0
			closed := false

			takeProfitPrice := activeShortTrade.EntryPrice * (1 - takeProfitPct)
			stopLossPrice := activeShortTrade.EntryPrice * (1 + stopLossPct)

			if currentCandle.Low <= takeProfitPrice {
				exitPrice = takeProfitPrice
				closed = true
			} else if currentCandle.High >= stopLossPrice {
				exitPrice = stopLossPrice
				closed = true
			} else if isShortStop {
				exitPrice = currentCandle.Close
				closed = true
			}

			if closed {
				activeShortTrade.ExitTime = currentCandle.Time
				activeShortTrade.ExitPrice = exitPrice
				activeShortTrade.Pnl = activeShortTrade.EntryPrice - activeShortTrade.ExitPrice
				activeShortTrade.PnlPercentage = (activeShortTrade.Pnl / activeShortTrade.EntryPrice) * 100
				completedTrades = append(completedTrades, *activeShortTrade)
				activeShortTrade = nil
			}
		} else {
			// Short Entry
			if i >= 14 && i >= config.VWZPeriod-1 && i >= config.ADXPeriod-1 {
				if isShortEntry {
					activeShortTrade = &Trade{
						EntryTime:       currentCandle.Time,
						EntryPrice:      currentCandle.Close,
						Direction:       "short",
						EntryIndicators: indicators,
					}
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
		if i < 14 {
			continue
		}
		if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
			continue
		}

		indicators := strategyData.createTechnicalIndicators(i, config)

		isLongEntry, _ := longCondition(indicators, config)
		isShortEntry, _ := shortCondition(indicators, config)

		if isLongEntry {
			signal := EntrySignal{
				Time:      strategyData.Candles[i].Time,
				Price:     strategyData.Candles[i].Close,
				Direction: "long",
			}
			signals = append(signals, signal)
		}
		if isShortEntry {
			signal := EntrySignal{
				Time:      strategyData.Candles[i].Time,
				Price:     strategyData.Candles[i].Close,
				Direction: "short",
			}
			signals = append(signals, signal)
		}
	}

	return signals
}
