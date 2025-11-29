package strategy

import (
	"go-backtesting/config"
	"go-backtesting/market"
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
	Size            float64
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
	var activeLongTrade *Trade
	var activeShortTrade *Trade
	var completedTrades []Trade

	takeProfitPct := config.TPRate // 1% take profit
	stopLossPct := config.SLRate   // 1% stop loss

	for i := range strategyData.Candles {
		currentCandle := strategyData.Candles[i]
		indicators := strategyData.createTechnicalIndicators(i, config)
		direction, entry, stop := DetermineEntrySignal(indicators, config, longCondition, shortCondition)

		// --- Hedge Mode Logic ---
		if config.HedgeMode && stop {
			activeLongTrade, activeShortTrade = handleStopSignal(config, activeLongTrade, activeShortTrade, direction, currentCandle, indicators)
		}

		// --- 1. Exit Logic ---
		if activeLongTrade != nil && activeShortTrade != nil { // Hedged position exit logic
			longPnl := (currentCandle.Close - activeLongTrade.EntryPrice) * activeLongTrade.Size
			shortPnl := (activeShortTrade.EntryPrice - currentCandle.Close) * activeShortTrade.Size
			if longPnl+shortPnl > 0 {
				// Close both positions
				activeLongTrade.ExitTime = currentCandle.Time
				activeLongTrade.ExitPrice = currentCandle.Close
				activeLongTrade.Pnl = longPnl
				completedTrades = append(completedTrades, *activeLongTrade)
				activeLongTrade = nil

				activeShortTrade.ExitTime = currentCandle.Time
				activeShortTrade.ExitPrice = currentCandle.Close
				activeShortTrade.Pnl = shortPnl
				completedTrades = append(completedTrades, *activeShortTrade)
				activeShortTrade = nil
			}
		} else { // Single position exit logic
			if activeLongTrade != nil {
				if currentCandle.Close > activeLongTrade.EntryPrice*(1+takeProfitPct) ||
					currentCandle.Close < activeLongTrade.EntryPrice*(1-stopLossPct) {
					exitPrice := currentCandle.Close
					activeLongTrade.ExitTime = currentCandle.Time
					activeLongTrade.ExitPrice = exitPrice
					activeLongTrade.Pnl = (activeLongTrade.ExitPrice - activeLongTrade.EntryPrice) * activeLongTrade.Size
					activeLongTrade.PnlPercentage = (activeLongTrade.Pnl / (activeLongTrade.EntryPrice * activeLongTrade.Size)) * 100
					completedTrades = append(completedTrades, *activeLongTrade)
					activeLongTrade = nil // Close the position
				}
			}
			if activeShortTrade != nil {
				if currentCandle.Close < activeShortTrade.EntryPrice*(1-takeProfitPct) ||
					currentCandle.Close > activeShortTrade.EntryPrice*(1+stopLossPct) {
					exitPrice := currentCandle.Close
					activeShortTrade.ExitTime = currentCandle.Time
					activeShortTrade.ExitPrice = exitPrice
					activeShortTrade.Pnl = (activeShortTrade.EntryPrice - activeShortTrade.ExitPrice) * activeShortTrade.Size
					activeShortTrade.PnlPercentage = (activeShortTrade.Pnl / (activeShortTrade.EntryPrice * activeShortTrade.Size)) * 100
					completedTrades = append(completedTrades, *activeShortTrade)
					activeShortTrade = nil // Close the position
				}
			}
		}

		// --- Minimum Position Size Check ---
		if activeLongTrade != nil && activeLongTrade.Size < config.MinPositionSize {
			// Close both positions
			if activeShortTrade != nil {
				activeShortTrade.ExitTime = currentCandle.Time
				activeShortTrade.ExitPrice = currentCandle.Close
				completedTrades = append(completedTrades, *activeShortTrade)
				activeShortTrade = nil
			}
			activeLongTrade.ExitTime = currentCandle.Time
			activeLongTrade.ExitPrice = currentCandle.Close
			completedTrades = append(completedTrades, *activeLongTrade)
			activeLongTrade = nil
		}
		if activeShortTrade != nil && activeShortTrade.Size < config.MinPositionSize {
			// Close both positions
			if activeLongTrade != nil {
				activeLongTrade.ExitTime = currentCandle.Time
				activeLongTrade.ExitPrice = currentCandle.Close
				completedTrades = append(completedTrades, *activeLongTrade)
				activeLongTrade = nil
			}
			activeShortTrade.ExitTime = currentCandle.Time
			activeShortTrade.ExitPrice = currentCandle.Close
			completedTrades = append(completedTrades, *activeShortTrade)
			activeShortTrade = nil
		}

		// --- 2. Entry Logic: Only enter if there is no active trade of the same direction ---
		if i < 14 { // Hardcoded ATR period in DetectBBWState
			continue
		}
		if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
			continue
		}

		if entry {
			if direction == "long" && activeLongTrade == nil {
				activeLongTrade = &Trade{
					EntryTime:       currentCandle.Time,
					EntryPrice:      currentCandle.Close,
					Direction:       direction,
					Size:            1.0,
					EntryIndicators: indicators,
				}
			} else if direction == "short" && activeShortTrade == nil {
				activeShortTrade = &Trade{
					EntryTime:       currentCandle.Time,
					EntryPrice:      currentCandle.Close,
					Direction:       direction,
					Size:            1.0,
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

func handleStopSignal(config *config.Config, activeLongTrade, activeShortTrade *Trade, direction string, currentCandle market.Candle, indicators TechnicalIndicators) (*Trade, *Trade) {
	if direction == "long" && activeLongTrade != nil {
		if activeShortTrade == nil {
			activeShortTrade = &Trade{
				EntryTime:       currentCandle.Time,
				EntryPrice:      currentCandle.Close,
				Direction:       "short",
				Size:            activeLongTrade.Size * config.HedgeSizeMultiplier,
				EntryIndicators: indicators,
			}
		} else {
			activeShortTrade.Size = activeLongTrade.Size / 2
		}
	} else if direction == "short" && activeShortTrade != nil {
		if activeLongTrade == nil {
			activeLongTrade = &Trade{
				EntryTime:       currentCandle.Time,
				EntryPrice:      currentCandle.Close,
				Direction:       "long",
				Size:            activeShortTrade.Size * config.HedgeSizeMultiplier,
				EntryIndicators: indicators,
			}
		} else {
			activeLongTrade.Size = activeShortTrade.Size / 2
		}
	}
	return activeLongTrade, activeShortTrade
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
