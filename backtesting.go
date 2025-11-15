package main

import (
	"fmt"
	"time"
)

// Trade는 단일 거래의 모든 세부 정보를 담는 구조체입니다.
type Trade struct {
	EntryTime       time.Time
	EntryPrice      float64
	ExitTime        time.Time
	ExitPrice       float64
	Direction       string // "long" or "short"
	Pnl             float64
	PnlPercentage   float64
	EntryIndicators TechnicalIndicators // 진입 시점의 모든 기술 지표
}

// BacktestResult는 백테스트 실행 후의 전체 결과를 담는 구조체입니다.
type BacktestResult struct {
	Trades      []Trade
	TotalPnl    float64
	WinCount    int
	LossCount   int
	TotalTrades int
	WinRate     float64
}

// runBacktest는 주어진 데이터와 설정으로 백테스팅을 실행하고 결과를 반환합니다.
func runBacktest(strategyData *StrategyDataContext, config *Config) BacktestResult {
	var activeTrade *Trade
	var completedTrades []Trade

	takeProfitPct := 0.01 // 1% 익절
	stopLossPct := 0.01   // 1% 손절

	for i := range strategyData.Candles {
		currentCandle := strategyData.Candles[i]

		// --- 1. Exit 로직: 활성화된 거래가 있는지 확인 ---
		if activeTrade != nil {
			var exitPrice float64
			exitTriggered := false

			if activeTrade.Direction == "long" {
				takeProfitPrice := activeTrade.EntryPrice * (1 + takeProfitPct)
				stopLossPrice := activeTrade.EntryPrice * (1 - stopLossPct)
				if currentCandle.High >= takeProfitPrice {
					exitPrice = takeProfitPrice
					exitTriggered = true
				} else if currentCandle.Low <= stopLossPrice {
					exitPrice = stopLossPrice
					exitTriggered = true
				}
			} else { // short
				takeProfitPrice := activeTrade.EntryPrice * (1 - takeProfitPct)
				stopLossPrice := activeTrade.EntryPrice * (1 + stopLossPct)
				if currentCandle.Low <= takeProfitPrice {
					exitPrice = takeProfitPrice
					exitTriggered = true
				} else if currentCandle.High >= stopLossPrice {
					exitPrice = stopLossPrice
					exitTriggered = true
				}
			}

			if exitTriggered {
				activeTrade.ExitTime = currentCandle.Time
				activeTrade.ExitPrice = exitPrice
				if activeTrade.Direction == "long" {
					activeTrade.Pnl = activeTrade.ExitPrice - activeTrade.EntryPrice
				} else {
					activeTrade.Pnl = activeTrade.EntryPrice - activeTrade.ExitPrice
				}
				activeTrade.PnlPercentage = (activeTrade.Pnl / activeTrade.EntryPrice) * 100
				completedTrades = append(completedTrades, *activeTrade)
				activeTrade = nil // 포지션 종료
			}
		}

		// --- 2. Entry 로직: 활성화된 거래가 없는 경우에만 진입 ---
		if activeTrade == nil {
			if i < config.VWZPeriod-1 || i < config.ADXPeriod-1 {
				continue
			}
			indicators := strategyData.createTechnicalIndicators(i)
			direction, hasSignal := determineEntrySignal(indicators, config.ADXThreshold)

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

	// --- 3. 최종 결과 계산 ---
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

// printBacktestSummary는 백테스트 결과 요약을 콘솔에 출력합니다.
func printBacktestSummary(result BacktestResult) {
	fmt.Printf("\n--- Backtest Summary ---\n")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("Total Trades: %d\n", result.TotalTrades)
	fmt.Printf("Win Rate: %.2f%%\n", result.WinRate)
	fmt.Printf("Wins: %d\n", result.WinCount)
	fmt.Printf("Losses: %d\n", result.LossCount)
	// fmt.Printf("Total PnL: %.2f\n", result.TotalPnl) // PnL is in price points, not currency
	fmt.Println("-----------------------------------------------------------------")
}
