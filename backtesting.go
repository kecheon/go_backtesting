package main

import (
	"fmt"
	"time"
)

// Trade는 단일 거래의 모든 세부 정보를 담는 구조체입니다.
type Trade struct {
	EntryTime                time.Time
	EntryPrice               float64
	ExitTime                 time.Time
	ExitPrice                float64
	Direction                string // "long" or "short"
	Pnl                      float64
	PnlPercentage            float64
	EntryIndicators          TechnicalIndicators // 진입 시점의 모든 기술 지표
	IsPriceThresholdBreached bool                // 익절/손절 가격에 한 번이라도 도달했는지 여부
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

	takeProfitPct := 0.02 // 1% 익절
	stopLossPct := 0.01   // 1% 손절

	for i := range strategyData.Candles {
		currentCandle := strategyData.Candles[i]

		// --- 1. Exit 로직: 활성화된 거래가 있는지 확인 ---
		if activeTrade != nil {
			// 1.1: 가격이 익절/손절선을 넘었는지 확인하고 상태 업데이트
			if !activeTrade.IsPriceThresholdBreached {
				if activeTrade.Direction == "long" {
					takeProfitPrice := activeTrade.EntryPrice * (1 + takeProfitPct)
					stopLossPrice := activeTrade.EntryPrice * (1 - stopLossPct)
					if currentCandle.High >= takeProfitPrice || currentCandle.Low <= stopLossPrice {
						activeTrade.IsPriceThresholdBreached = true
					}
				} else { // short
					takeProfitPrice := activeTrade.EntryPrice * (1 - takeProfitPct)
					stopLossPrice := activeTrade.EntryPrice * (1 + stopLossPct)
					if currentCandle.Low <= takeProfitPrice || currentCandle.High >= stopLossPrice {
						activeTrade.IsPriceThresholdBreached = true
					}
				}
			}

			// 1.2: 최종 종료 조건 확인 (가격 조건이 만족된 상태에서 DMI 조건이 허락하는가?)
			finalExitTrigger := false
			if activeTrade.IsPriceThresholdBreached {
				shouldHold := false
				if activeTrade.Direction == "long" &&
					i < len(strategyData.PlusDI) &&
					i < len(strategyData.MinusDI) &&
					strategyData.PlusDI[i] > strategyData.MinusDI[i] {
					shouldHold = true
				} else if activeTrade.Direction == "short" &&
					i < len(strategyData.PlusDI) &&
					i < len(strategyData.MinusDI) &&
					strategyData.MinusDI[i] > strategyData.PlusDI[i] {
					shouldHold = true
				}

				if !shouldHold {
					finalExitTrigger = true
				}
			}

			// 1.3: 종료 실행
			if finalExitTrigger {
				// 실제 종료가 일어나는 캔들의 종가를 Exit Price로 사용
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

// printDetailedTradeRecords는 모든 완료된 거래 기록의 상세 내용을 콘솔에 출력합니다.
func printDetailedTradeRecords(result BacktestResult) {
	fmt.Printf("\n--- Detailed Trade Records ---\n")
	fmt.Println("-----------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-5s %-5s %-20s %-15s %-20s %-15s %-10s %-10s %-10s\n",
		"Idx", "Type", "Entry Time", "Entry Price", "Exit Time", "Exit Price", "Pnl", "Pnl(%)", "Status")
	fmt.Println("-----------------------------------------------------------------------------------------------------------------------------------------")

	for i, trade := range result.Trades {
		status := "Loss"
		if trade.Pnl > 0 {
			status = "Win"
		}

		fmt.Printf("%-5d %-5s %-20s %-15.2f %-20s %-15.2f %-10.2f %-9.2f%% %-10s\n",
			i,
			trade.Direction,
			trade.EntryTime.Format("01-02 15:04:05"),
			trade.EntryPrice,
			trade.ExitTime.Format("01-02 15:04:05"),
			trade.ExitPrice,
			trade.Pnl,
			trade.PnlPercentage,
			status,
		)
	}
	fmt.Println("-----------------------------------------------------------------------------------------------------------------------------------------")
}

// printBacktestSummary는 백테스트 결과 요약을 콘솔에 출력합니다.
func printBacktestSummary(result BacktestResult) {
	fmt.Printf("\n--- Backtest Summary ---\n")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("Total Trades: %d\n", result.TotalTrades)
	fmt.Printf("Win Rate: %.2f%%\n", result.WinRate)
	fmt.Printf("Wins: %d\n", result.WinCount)
	fmt.Printf("Losses: %d\n", result.LossCount)
	fmt.Printf("Total PnL: %.2f\n", result.TotalPnl) // PnL is in price points, not currency
	fmt.Println("-----------------------------------------------------------------")
}
