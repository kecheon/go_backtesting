package reporting_test

import (
	"go-backtesting/reporting"
	"go-backtesting/strategy"
	"testing"
	"time"
)

func TestPrintTradeAnalysis(t *testing.T) {
	result := strategy.BacktestResult{
		Trades: []strategy.Trade{
			{
				EntryTime:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				EntryPrice:    100,
				ExitTime:      time.Date(2023, 1, 1, 0, 5, 0, 0, time.UTC),
				ExitPrice:     105,
				Direction:     "long",
				Pnl:           5,
				PnlPercentage: 5,
				EntryIndicators: strategy.TechnicalIndicators{
					ZScore:   []float64{1.0},
					VWZScore: []float64{1.0},
					ADX:      []float64{25.0},
					PlusDI:   []float64{20.0},
					MinusDI:  []float64{15.0},
					DX:       []float64{10.0},
				},
			},
		},
	}
	strategyData := &strategy.StrategyDataContext{}

	// This is a smoke test to ensure the function doesn't panic.
	// A more thorough test would capture the output and verify it.
	reporting.PrintTradeAnalysis(result, strategyData)
}
