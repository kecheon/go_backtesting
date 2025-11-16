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
					ZScore:   1.0,
					VWZScore: 1.0,
					ADX:      25.0,
					PlusDI:   20.0,
					MinusDI:  15.0,
					DX:       10.0,
				},
			},
		},
	}
	strategyData := &strategy.StrategyDataContext{}

	// This is a smoke test to ensure the function doesn't panic.
	// A more thorough test would capture the output and verify it.
	reporting.PrintTradeAnalysis(result, strategyData)
}
