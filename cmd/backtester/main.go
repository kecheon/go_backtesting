package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"go-backtesting/internal/backtest"
	"go-backtesting/internal/model"
	"go-backtesting/internal/reporting"
	"strconv"
)

func main() {
	// 1. Load Data
	candles, err := loadCandles("btc_1h.csv")
	if err != nil {
		fmt.Println("Error loading data:", err)
		return
	}
	fmt.Printf("Loaded %d candles\n", len(candles))

	// 2. Configure Strategy
	cfg := backtest.Config{
		LookbackPeriod: 240,   // 10 days of hours
		BinSizePct:     0.05,  // 0.05% bin size
		ValueAreaPct:   0.70,  // 70% VA
		POCProximity:   0.2,   // 0.2% proximity
	}

	// 3. Run Backtest
	engine := backtest.NewEngine(candles, cfg)
	fmt.Println("Running backtest...")
	engine.Run()

	// 4. Report Results
	stats := reporting.CalculateStats(engine.Trades)
	fmt.Printf("\n=== Backtest Results ===\n")
	fmt.Printf("Total Trades: %d\n", stats.TotalTrades)
	fmt.Printf("Win Rate:     %.2f%%\n", stats.WinRate)
	fmt.Printf("Profit Factor: %.2f\n", stats.ProfitFactor)
	fmt.Printf("Max Drawdown:  %.2f\n", stats.MaxDrawdown)
	fmt.Printf("Total PnL:     %.2f\n", stats.TotalPnL)

	// Save Logs
	saveTradeLog(engine.Trades)

	// Save Chart
	err = reporting.GenerateHTMLChart(candles, engine.Trades, "chart.html")
	if err != nil {
		fmt.Println("Error saving chart:", err)
	} else {
		fmt.Println("Chart saved to chart.html")
	}
}

func loadCandles(path string) ([]model.Candle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var candles []model.Candle
	for i, row := range records {
		if i == 0 { continue } // Header

		ts, _ := strconv.ParseInt(row[0], 10, 64)
		o, _ := strconv.ParseFloat(row[1], 64)
		h, _ := strconv.ParseFloat(row[2], 64)
		l, _ := strconv.ParseFloat(row[3], 64)
		c, _ := strconv.ParseFloat(row[4], 64)
		v, _ := strconv.ParseFloat(row[5], 64)

		candles = append(candles, model.Candle{
			Timestamp: ts,
			Open:      o,
			High:      h,
			Low:       l,
			Close:     c,
			Volume:    v,
		})
	}
	return candles, nil
}

func saveTradeLog(trades []*model.Trade) {
	f, _ := os.Create("trades.csv")
	defer f.Close()

	fmt.Fprintln(f, "ID,Side,EntryTime,EntryPrice,ExitTime,ExitPrice,PnL,Reason,Pattern")
	for _, t := range trades {
		fmt.Fprintf(f, "%d,%s,%d,%.2f,%d,%.2f,%.2f,%s,%s\n",
			t.ID, t.Side, t.EntryTime, t.EntryPrice, t.ExitTime, t.ExitPrice, t.PnL, t.Reason, t.EntryPattern)
	}
}
